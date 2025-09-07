// pkg/core/response.go

package core

// Response 响应接口
type Response interface {
	MsgType() string
	RequestID() uint64
	Bind(target interface{}) error
	RawData() []byte
}

// response 响应实现
type response struct {
	msgType   string
	requestID uint64
	rawData   []byte
	processor Processor
}

func (r *response) MsgType() string {
	return r.msgType
}

func (r *response) RequestID() uint64 {
	return r.requestID
}

func (r *response) Bind(target interface{}) error {
	return r.processor.Serializer().Deserialize(r.rawData, target)
}

func (r *response) RawData() []byte {
	return r.rawData
}
