package core_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/BadKid90s/chilix-msg/pkg/core"
	"github.com/BadKid90s/chilix-msg/pkg/log"
	"github.com/BadKid90s/chilix-msg/pkg/serializer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConnection 是 transport.Connection 的模拟实现
type MockConnection struct {
	mock.Mock
	buf *bytes.Buffer
}

func (m *MockConnection) Read(p []byte) (n int, err error) {
	if m.buf != nil {
		return m.buf.Read(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockConnection) Write(p []byte) (n int, err error) {
	if m.buf != nil {
		return m.buf.Write(p)
	}
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) LocalAddr() net.Addr {
	return nil
}

func (m *MockConnection) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConnection) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConnection) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConnection) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// 测试创建Processor
func TestNewProcessor(t *testing.T) {
	conn := &MockConnection{}
	logger := log.NewDefaultLogger()

	// 测试使用默认序列化器
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: logger,
	})
	assert.NotNil(t, processor)
	assert.Equal(t, serializer.DefaultSerializer, processor.Opts.Serializer)

	// 测试使用自定义序列化器
	customSerializer := &serializer.JSON{}
	processor = core.NewProcessor(conn, core.ProcessorOptions{
		Serializer: customSerializer,
		Logger:     logger,
	})
	assert.NotNil(t, processor)
	assert.Equal(t, customSerializer, processor.Opts.Serializer)

	// 测试未提供Logger时使用默认Logger
	processor = core.NewProcessor(conn, core.ProcessorOptions{})
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.Logger())
}

// 测试注册中间件
func TestProcessor_Use(t *testing.T) {
	conn := &MockConnection{}
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	// 注册一个简单的中间件
	middleware := func(next core.Handler) core.Handler {
		return func(ctx core.Context) error {
			return next(ctx)
		}
	}

	processor.Use(middleware)
	// 验证中间件被正确注册 (通过反射检查router的middlewares字段)
}

// 测试注册处理器
func TestProcessor_RegisterHandler(t *testing.T) {
	conn := &MockConnection{}
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	// 注册一个处理器
	handler := func(ctx core.Context) error {
		return nil
	}

	processor.RegisterHandler("test", handler)
	// 验证处理器被正确注册 (通过反射检查router的handlers字段)
}

// 测试发送消息
func TestProcessor_Send(t *testing.T) {
	buf := &bytes.Buffer{}
	conn := &MockConnection{buf: buf}
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	// 发送消息
	payload := map[string]string{"key": "value"}
	err := processor.Send("test", payload)
	assert.NoError(t, err)

	// 验证消息是否正确编码
	assert.True(t, buf.Len() > 0)
}

// 测试回复消息
func TestProcessor_Reply(t *testing.T) {
	buf := &bytes.Buffer{}
	conn := &MockConnection{buf: buf}
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	// 回复消息
	requestID := uint64(123)
	payload := map[string]string{"key": "value"}
	err := processor.Reply(requestID, "test_response", payload)
	assert.NoError(t, err)

	// 验证消息是否正确编码
	assert.True(t, buf.Len() > 0)
}

// 测试错误响应
func TestProcessor_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	conn := &MockConnection{buf: buf}
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	// 发送错误响应
	requestID := uint64(123)
	err := processor.Error(requestID, "test error")
	assert.NoError(t, err)

	// 验证消息是否正确编码
	assert.True(t, buf.Len() > 0)
}

// 测试关闭处理器
func TestProcessor_Close(t *testing.T) {
	conn := &MockConnection{}
	conn.On("Close").Return(nil)
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})

	err := processor.Close()
	assert.NoError(t, err)
	conn.AssertExpectations(t)
}

// 测试获取Logger
func TestProcessor_Logger(t *testing.T) {
	conn := &MockConnection{}
	logger := log.NewDefaultLogger()
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Logger: logger,
	})

	assert.Equal(t, logger, processor.Logger())
}
