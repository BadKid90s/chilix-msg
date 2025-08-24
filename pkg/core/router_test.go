package core

import (
	"testing"

	"github.com/BadKid90s/chilix-msg/pkg/log"
	"github.com/BadKid90s/chilix-msg/pkg/transport"
	"github.com/stretchr/testify/assert"
)

// MockContext 是 Context 接口的模拟实现
type MockContext struct {
	msgType string
}

func (m *MockContext) Logger() log.Logger {
	return nil
}

func (m *MockContext) Bind(interface{}) error {
	return nil
}

func (m *MockContext) MessageType() string {
	return m.msgType
}

func (m *MockContext) RequestID() uint64 {
	return 0
}

func (m *MockContext) IsRequest() bool {
	return false
}

func (m *MockContext) IsResponse() bool {
	return true
}

func (m *MockContext) Connection() transport.Connection {
	return nil
}

func (m *MockContext) RawData() []byte {
	return nil
}

func (m *MockContext) SetRawData([]byte) {
}

func (m *MockContext) Writer() interface{} {
	return nil
}

func (m *MockContext) SetWriter(Writer) {
}

func (m *MockContext) Error(string) error {
	return nil
}

func (m *MockContext) Processor() *Processor {
	return nil
}

func (m *MockContext) Reply(interface{}) error {
	return nil
}

// 测试创建Router
func TestNewRouter(t *testing.T) {
	router := NewRouter()
	assert.NotNil(t, router)
	assert.NotNil(t, router.Handlers())
	assert.Empty(t, router.Handlers())
	assert.Empty(t, router.Middlewares())
}

// 测试注册中间件
func TestRouter_Use(t *testing.T) {
	router := NewRouter()

	middleware := func(next Handler) Handler {
		return func(ctx Context) error {
			return next(ctx)
		}
	}

	router.Use(middleware)
	assert.Equal(t, 1, len(router.Middlewares()))
}

// 测试注册处理器
func TestRouter_Handle(t *testing.T) {
	router := NewRouter()

	handler := func(ctx Context) error {
		return nil
	}

	router.Handle("test", handler)
	assert.Contains(t, router.Handlers(), "test")
}

// 测试分发消息
func TestRouter_Dispatch(t *testing.T) {
	router := NewRouter()

	// 注册处理器
	called := false
	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Handle("test", handler)

	// 创建模拟上下文
	ctx := &MockContext{msgType: "test"}

	// 分发消息
	err := router.Dispatch("test", ctx)
	assert.NoError(t, err)
	assert.True(t, called)

	// 测试不存在的消息类型
	err = router.Dispatch("unknown", ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrHandlerNotFound, err)
}
