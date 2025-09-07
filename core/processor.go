package core

import (
	"time"

	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
)

// Middleware 中间件类型
type Middleware func(Handler) Handler

// Handler 消息处理器类型
type Handler func(ctx Context) error

// Processor 消息处理器接口 - 用户面向的简洁接口
type Processor interface {
	// RegisterHandler 消息处理
	RegisterHandler(msgType string, handler Handler) error
	// Use 中间件
	Use(middleware Middleware)

	// Send 消息发送
	Send(msgType string, payload interface{}) error
	// Request 带请求ID的消息发送
	Request(msgType string, payload interface{}) (Response, error)
	// Reply 回复消息
	Reply(requestID uint64, msgType string, payload interface{}) error

	// Listen 生命周期管理
	Listen() error
	// Close 销毁
	Close() error

	// Logger 配置访问
	Logger() log.Logger
	// Serializer 配置序列化方式
	Serializer() serializer.Serializer
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	Serializer       serializer.Serializer // 序列化器
	MessageSizeLimit int64                 // 消息大小限制（字节）
	RequestTimeout   time.Duration         // 请求超时时间
	Logger           log.Logger            // 日志记录器
}

// NewProcessor 创建新的消息处理器
func NewProcessor(conn transport.Connection, config ProcessorConfig) Processor {
	return newProcessor(conn, config)
}
