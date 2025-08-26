package core

// Writer 消息写入器接口
type Writer interface {
	// Write 发送消息
	Write(msgType string, payload interface{}) error

	// Reply 发送响应
	Reply(requestID uint64, msgType string, payload interface{}) error

	// Error 发送错误响应
	Error(errorMsg string) error
}

// messageWriter 消息写入器实现
type messageWriter struct {
	processor *Processor
	requestID uint64  // 当前请求ID，用于错误响应
}

func NewMessageWriter(p *Processor) Writer {
	return &messageWriter{processor: p}
}

// NewMessageWriterWithRequest 为请求上下文创建消息写入器
func NewMessageWriterWithRequest(p *Processor, requestID uint64) Writer {
	return &messageWriter{processor: p, requestID: requestID}
}

func (w *messageWriter) Write(msgType string, payload interface{}) error {
	return w.processor.Send(msgType, payload)
}

func (w *messageWriter) Reply(requestID uint64, msgType string, payload interface{}) error {
	return w.processor.Reply(requestID, msgType, payload)
}

func (w *messageWriter) Error(errorMsg string) error {
	// 如果存在 requestID，发送错误响应
	if w.requestID > 0 {
		return w.processor.Reply(w.requestID, "error", map[string]string{"error": errorMsg})
	}
	// 否则发送错误消息
	return w.processor.Send("error", map[string]string{"error": errorMsg})
}
