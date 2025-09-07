package core

import (
	"sync/atomic"
)

// RequestIDGenerator 请求ID生成器
type RequestIDGenerator struct {
	counter uint64
}

// NewRequestIDGenerator 创建请求ID生成器
func NewRequestIDGenerator() *RequestIDGenerator {
	return &RequestIDGenerator{}
}

// Next 获取下一个Request ID
func (g *RequestIDGenerator) Next() uint64 {
	return atomic.AddUint64(&g.counter, 1)
}
