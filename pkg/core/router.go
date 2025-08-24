package core

import (
	"sync"
)

// Middleware 中间件类型
type Middleware func(Handler) Handler

// Handler 消息处理器类型
type Handler func(ctx Context) error

// Router 消息路由器
type Router struct {
	handlers    map[string]Handler
	middlewares []Middleware
	mutex       sync.RWMutex
}

// NewRouter 创建新路由器
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

// Use 注册中间件
func (r *Router) Use(middleware Middleware) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.middlewares = append(r.middlewares, middleware)
}

// Handle 注册处理器
func (r *Router) Handle(msgType string, handler Handler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 应用中间件
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	r.handlers[msgType] = handler
}

// Dispatch 分发消息
func (r *Router) Dispatch(msgType string, ctx Context) error {
	r.mutex.RLock()
	h, ok := r.handlers[msgType]
	r.mutex.RUnlock()

	if !ok {
		return ErrHandlerNotFound
	}

	return h(ctx)
}
