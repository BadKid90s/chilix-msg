package core

import (
	"sync"
	"time"
)

// RequestManager 请求管理器
type RequestManager struct {
	pending sync.Map
	idGen   *RequestIDGenerator
	timeout time.Duration
}

func NewRequestManager(timeout time.Duration) *RequestManager {
	return &RequestManager{
		idGen:   NewRequestIDGenerator(),
		timeout: timeout,
	}
}

// StartRequest 开始一个新请求
func (rm *RequestManager) StartRequest() (uint64, chan Response) {
	requestID := rm.idGen.Next()
	ch := make(chan Response, 1)
	rm.pending.Store(requestID, ch)
	return requestID, ch
}

// CompleteRequest 完成请求
func (rm *RequestManager) CompleteRequest(requestID uint64, response Response) {
	if ch, ok := rm.pending.Load(requestID); ok {
		ch.(chan Response) <- response
		rm.pending.Delete(requestID)
	}
}

// CancelRequest 取消请求
func (rm *RequestManager) CancelRequest(requestID uint64) {
	rm.pending.Delete(requestID)
}

// IsPending 检查请求是否在pending中
func (rm *RequestManager) IsPending(requestID uint64) (chan Response, bool) {
	ch, ok := rm.pending.Load(requestID)
	if !ok {
		return nil, false
	}
	return ch.(chan Response), true
}
