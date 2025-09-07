package core

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/BadKid90s/chilix-msg/codec"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
)

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrHandlerNotFound = errors.New("no handler for message type")
)

// processor 内部实现，不对外暴露
type processor struct {
	conn         transport.Connection
	codec        codec.Codec
	handlers     map[string]Handler
	middlewares  []Middleware
	typeRegistry *Registry
	requestMgr   *RequestManager
	config       ProcessorConfig
	ctx          context.Context
	cancel       context.CancelFunc
	logger       log.Logger
	serializer   serializer.Serializer
	mutex        sync.RWMutex
}

// newProcessor 创建新的处理器实例
func newProcessor(conn transport.Connection, config ProcessorConfig) *processor {
	if config.Serializer == nil {
		config.Serializer = serializer.DefaultSerializer
	}

	if config.Logger == nil {
		config.Logger = log.NewDefaultLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &processor{
		conn:         conn,
		codec:        codec.NewBalancedCodec(config.Serializer),
		handlers:     make(map[string]Handler),
		middlewares:  make([]Middleware, 0),
		typeRegistry: NewRegistry(),
		requestMgr:   NewRequestManager(config.RequestTimeout),
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		logger:       config.Logger,
		serializer:   config.Serializer,
	}
}

// RegisterHandler 注册消息处理器
func (p *processor) RegisterHandler(msgType string, handler Handler) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 注册类型到注册器
	_, err := p.typeRegistry.Register(msgType)
	if err != nil {
		return err
	}

	// 应用中间件
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		handler = p.middlewares[i](handler)
	}

	// 注册处理器函数
	p.handlers[msgType] = handler
	return nil
}

// Use 注册中间件
func (p *processor) Use(middleware Middleware) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.middlewares = append(p.middlewares, middleware)
}

// Listen 开始监听和处理消息
func (p *processor) Listen() error {
	for {
		select {
		case <-p.ctx.Done():
			return nil
		default:
			// 读取消息
			msgTypeID, rawData, requestID, err := p.codec.Decode(p.conn)
			if err != nil {
				p.logger.Errorf("Failed to decode message: %v", err)
				// 根据错误类型决定是否继续监听
				if p.isRecoverableError(err) {
					continue // 可恢复错误，继续监听
				}
				return err // 不可恢复错误，退出监听
			}

			// 将类型ID转换为类型字符串
			msgType, exists := p.typeRegistry.GetName(msgTypeID)
			if !exists {
				p.logger.Errorf("Unknown message type ID: %d", msgTypeID)
				continue
			}

			// 检查消息大小
			if p.config.MessageSizeLimit > 0 && int64(len(rawData)) > p.config.MessageSizeLimit {
				p.logger.Warnf("Message too large: %d > %d", len(rawData), p.config.MessageSizeLimit)
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
				if err := p.dispatchMessage(msgType, ctx); err != nil {
					p.logger.Errorf("Error processing message %s: %v", msgType, err)
				}
			}()
		}
	}
}

// dispatchMessage 分发消息到对应的处理器
func (p *processor) dispatchMessage(msgType string, ctx Context) error {
	p.mutex.RLock()
	handler, exists := p.handlers[msgType]
	p.mutex.RUnlock()

	if !exists {
		return ErrHandlerNotFound
	}

	return handler(ctx)
}

// Send 发送消息
func (p *processor) Send(msgType string, payload interface{}) error {
	p.logger.Debugf("Sending message: msgType=%s", msgType)

	// 获取类型ID
	msgTypeID, exists := p.typeRegistry.GetID(msgType)
	if !exists {
		// 如果类型未注册，先注册
		var err error
		msgTypeID, err = p.typeRegistry.Register(msgType)
		if err != nil {
			return err
		}
	}

	return p.codec.Encode(p.conn, msgTypeID, payload, 0)
}

// Request 发送请求并等待响应
func (p *processor) Request(msgType string, payload interface{}) (Response, error) {
	p.logger.Debugf("Sending request: msgType=%s", msgType)

	// 开始新请求
	requestID, ch := p.requestMgr.StartRequest()

	// 获取类型ID
	msgTypeID, exists := p.typeRegistry.GetID(msgType)
	if !exists {
		// 如果类型未注册，先注册
		var err error
		msgTypeID, err = p.typeRegistry.Register(msgType)
		if err != nil {
			p.requestMgr.CancelRequest(requestID)
			return nil, err
		}
	}

	// 发送请求
	if err := p.codec.Encode(p.conn, msgTypeID, payload, requestID); err != nil {
		p.requestMgr.CancelRequest(requestID)
		return nil, err
	}

	// 等待响应
	select {
	case response := <-ch:
		return response, nil
	case <-time.After(p.config.RequestTimeout):
		p.requestMgr.CancelRequest(requestID)
		return nil, ErrRequestTimeout
	}
}

// Reply 发送响应（内部方法）
func (p *processor) Reply(requestID uint64, msgType string, payload interface{}) error {
	p.logger.Debugf("Sending reply: requestID=%d, msgType=%s", requestID, msgType)

	// 获取类型ID
	msgTypeID, exists := p.typeRegistry.GetID(msgType)
	if !exists {
		// 如果类型未注册，先注册
		var err error
		msgTypeID, err = p.typeRegistry.Register(msgType)
		if err != nil {
			return err
		}
	}

	return p.codec.Encode(p.conn, msgTypeID, payload, requestID)
}

// Logger 返回配置的日志记录器
func (p *processor) Logger() log.Logger {
	return p.logger
}

// Serializer 返回序列化器
func (p *processor) Serializer() serializer.Serializer {
	return p.serializer
}

// Close 关闭处理器
func (p *processor) Close() error {
	p.cancel()
	p.logger.Infof("Processor closed")
	return p.conn.Close()
}

// isRecoverableError 判断错误是否可恢复
func (p *processor) isRecoverableError(err error) bool {
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
