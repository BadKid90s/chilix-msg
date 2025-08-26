package core

import (
	"context"
	"time"

	"github.com/BadKid90s/chilix-msg/codec"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
)

// ProcessorOptions 处理器配置选项
type ProcessorOptions struct {
	Serializer       serializer.Serializer // 序列化器
	MessageSizeLimit int64                 // 消息大小限制（字节）
	RequestTimeout   time.Duration         // 请求超时时间
	Logger           log.Logger            // 日志记录器
}

// Processor 消息处理器
type Processor struct {
	conn       transport.Connection
	codec      codec.Codec
	router     *Router
	opts       ProcessorOptions
	ctx        context.Context
	cancel     context.CancelFunc
	requestMgr *RequestManager
	logger     log.Logger
	serializer serializer.Serializer
}

// NewProcessor 创建新的处理器
func NewProcessor(conn transport.Connection, opts ProcessorOptions) *Processor {
	if opts.Serializer == nil {
		opts.Serializer = serializer.DefaultSerializer
	}

	// 如果没有配置日志记录器，则使用默认的日志记录器
	if opts.Logger == nil {
		opts.Logger = log.NewDefaultLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Processor{
		conn:       conn,
		codec:      codec.NewLengthPrefixCodec(opts.Serializer),
		router:     NewRouter(),
		opts:       opts,
		ctx:        ctx,
		cancel:     cancel,
		requestMgr: NewRequestManager(opts.RequestTimeout),
		serializer: opts.Serializer,
		logger:     opts.Logger,
	}
}

// Use 注册中间件
func (p *Processor) Use(middleware Middleware) {
	p.router.Use(middleware)
}

// RegisterHandler 注册消息处理器
func (p *Processor) RegisterHandler(msgType string, handler Handler) {
	// 注册处理器函数
	p.router.Handle(msgType, handler)
}

// Listen 开始监听和处理消息
func (p *Processor) Listen() error {
	for {
		select {
		case <-p.ctx.Done():
			return nil
		default:
			// 读取消息
			msgType, rawData, requestID, err := p.codec.Decode(p.conn)
			if err != nil {
				p.logger.Errorf("Failed to decode message: %v", err)
				// 根据错误类型决定是否继续监听
				if p.isRecoverableError(err) {
					continue // 可恢复错误，继续监听
				}
				return err // 不可恢复错误，退出监听
			}

			// 检查消息大小
			if p.opts.MessageSizeLimit > 0 && int64(len(rawData)) > p.opts.MessageSizeLimit {
				p.logger.Warnf("Message too large: %d > %d", len(rawData), p.opts.MessageSizeLimit)
				continue
			}

			// 如果是响应消息（RequestID > 0 且 pending 中有匹配项）
			if requestID > 0 {
				if ch, ok := p.requestMgr.IsPending(requestID); ok {

					// 完成请求
					response := &response{
						msgType:   msgType,
						requestID: requestID,
						rawData:   rawData,
						processor: p,
					}
					ch <- response
					p.requestMgr.CancelRequest(requestID)
					continue
				}
			}

			// 创建上下文
			ctx := &processorContext{
				msgType:    msgType,
				requestID:  requestID,
				connection: p.conn,
				rawData:    rawData,
				processor:  p,
				writer:     NewMessageWriter(p),
				logger:     p.logger,
			}

			// 处理消息
			go func() {
				p.logger.Debugf("Dispatching message: msgType=%s, requestID=%d", msgType, requestID)
				if err := p.router.Dispatch(msgType, ctx); err != nil {
					p.logger.Errorf("Error processing message %s: %v", msgType, err)
				}
			}()
		}
	}
}

// Send 发送消息
func (p *Processor) Send(msgType string, payload interface{}) error {
	p.logger.Debugf("Sending message: msgType=%s", msgType)
	return p.codec.Encode(p.conn, msgType, payload, 0)
}

// Request 发送请求并等待响应
func (p *Processor) Request(msgType string, payload interface{}) (Response, error) {
	p.logger.Debugf("Sending request: msgType=%s", msgType)

	// 开始新请求
	requestID, ch := p.requestMgr.StartRequest()

	// 发送请求
	if err := p.codec.Encode(p.conn, msgType, payload, requestID); err != nil {
		p.requestMgr.CancelRequest(requestID)
		return nil, err
	}

	// 等待响应
	select {
	case response := <-ch:
		return response, nil
	case <-time.After(p.opts.RequestTimeout):
		p.requestMgr.CancelRequest(requestID)
		return nil, ErrRequestTimeout
	}
}

// Reply 发送响应
func (p *Processor) Reply(requestID uint64, msgType string, payload interface{}) error {
	p.logger.Debugf("Sending reply: requestID=%d, msgType=%s", requestID, msgType)
	return p.codec.Encode(p.conn, msgType, payload, requestID)
}

// Logger 返回配置的日志记录器
func (p *Processor) Logger() log.Logger {
	return p.logger
}
func (p *Processor) Serializer() serializer.Serializer {
	return p.serializer
}

// Close 关闭处理器
func (p *Processor) Close() error {
	p.cancel()
	p.logger.Infof("Processor closed")
	return p.conn.Close()
}

// isRecoverableError 判断错误是否可恢复
func (p *Processor) isRecoverableError(err error) bool {
	// 这里可以根据具体的错误类型来判断
	// 例如：网络临时错误、序列化错误等可以恢复
	// 连接关闭、严重协议错误等不可恢复
	switch err.Error() {
	case "EOF", "connection reset by peer":
		return false // 连接已关闭，不可恢复
	default:
		return true // 其他错误暂时视为可恢复
	}
}
