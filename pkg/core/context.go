package core

import "github.com/BadKid90s/chilix-msg/pkg/transport"

// Context 处理器上下文接口
type Context interface {
	Bind(target interface{}) error    // 绑定消息负载
	MessageType() string              // 获取消息类型
	RequestID() uint64                // 获取请求ID
	IsRequest() bool                  // 判断是否是请求消息
	IsResponse() bool                 // 判断是否是响应消息
	Connection() transport.Connection // 获取底层连接
	RawData() []byte                  // 获取原始数据
	SetRawData(data []byte)           // 设置原始数据
	//Writer() Writer                   // 获取写入器
	SetWriter(writer Writer)         // 设置写入器
	Reply(payload interface{}) error // 发送成功响应
	Error(errorMsg string) error     // 发送错误响应
}

// processorContext 处理器上下文实现
type processorContext struct {
	msgType    string
	requestID  uint64
	connection transport.Connection
	rawData    []byte
	processor  *Processor
	writer     Writer
}

func (c *processorContext) Bind(target interface{}) error {
	return c.processor.opts.Serializer.Deserialize(c.rawData, target)
}

func (c *processorContext) MessageType() string {
	return c.msgType
}

func (c *processorContext) RequestID() uint64 {
	return c.requestID
}

func (c *processorContext) IsRequest() bool {
	return c.requestID > 0
}

func (c *processorContext) IsResponse() bool {
	return !c.IsRequest()
}

func (c *processorContext) Connection() transport.Connection {
	return c.connection
}

func (c *processorContext) RawData() []byte {
	return c.rawData
}

func (c *processorContext) SetRawData(data []byte) {
	c.rawData = data
}

func (c *processorContext) SetWriter(writer Writer) {
	c.writer = writer
}

func (c *processorContext) Reply(payload interface{}) error {
	return c.writer.Reply(c.requestID, c.msgType, payload)
}

func (c *processorContext) Error(errorMsg string) error {
	return c.writer.Error(errorMsg)
}
