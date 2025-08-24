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
}

func NewMessageWriter(p *Processor) Writer {
	return &messageWriter{processor: p}
}

func (w *messageWriter) Write(msgType string, payload interface{}) error {
	return w.processor.Send(msgType, payload)
}

func (w *messageWriter) Reply(requestID uint64, msgType string, payload interface{}) error {
	return w.processor.Reply(requestID, msgType, payload)
}

func (w *messageWriter) Error(errorMsg string) error {
	return w.processor.Send("error", map[string]string{"error": errorMsg})
}
