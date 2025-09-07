package middleware

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/BadKid90s/chilix-msg/core"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
	"github.com/stretchr/testify/assert"
)

// MockConnection 是 transport.Connection 的模拟实现
type MockConnection struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func NewMockConnection() *MockConnection {
	return &MockConnection{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
	}
}

func (m *MockConnection) Read(p []byte) (n int, err error) {
	return m.readBuf.Read(p)
}

func (m *MockConnection) Write(p []byte) (n int, err error) {
	return m.writeBuf.Write(p)
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) LocalAddr() net.Addr {
	return nil
}

func (m *MockConnection) RemoteAddr() net.Addr {
	return nil

}

func (m *MockConnection) SetDeadline(time.Time) error {
	return nil
}

func (m *MockConnection) SetReadDeadline(time.Time) error {
	return nil
}

func (m *MockConnection) SetWriteDeadline(time.Time) error {
	return nil
}

// MockWriter 是 Writer 接口的模拟实现
type MockWriter struct {
	lastMsgType string
	lastPayload interface{}
	lastError   error
}

func (m *MockWriter) Write(msgType string, payload interface{}) error {
	m.lastMsgType = msgType
	m.lastPayload = payload
	return nil
}

func (m *MockWriter) Reply(_ uint64, msgType string, payload interface{}) error {
	m.lastMsgType = msgType
	m.lastPayload = payload
	return nil
}

func (m *MockWriter) Error(errorMsg string) error {
	m.lastError = fmt.Errorf("%s", errorMsg)
	return nil
}

// 测试加密和解密功能
func TestEncryptDecrypt(t *testing.T) {
	// 生成测试密钥
	key := KeyFromString("test-secret-key")

	// 测试数据
	originalData := []byte("This is a secret message for testing")

	// 加密数据
	encryptedData, err := encrypt(key, originalData)
	assert.NoError(t, err)
	assert.NotEqual(t, originalData, encryptedData)

	// 解密数据
	decryptedData, err := decrypt(key, encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

// 测试密钥生成函数
func TestKeyGeneration(t *testing.T) {
	// 测试从字符串生成密钥
	key1 := KeyFromString("password123")
	assert.Equal(t, 32, len(key1)) // SHA-256 produces 32-byte hash

	// 测试相同地输入产生相同的输出
	key2 := KeyFromString("password123")
	assert.Equal(t, key1, key2)

	// 测试不同的输入产生不同的输出
	key3 := KeyFromString("password124")
	assert.NotEqual(t, key1, key3)

	// 测试从Base64生成密钥
	key4, err := KeyFromBase64("dGVzdC1zZWNyZXQta2V5")
	assert.NoError(t, err)
	assert.NotEmpty(t, key4)
}

// 测试加密中间件
func TestEncryptionMiddleware(t *testing.T) {
	// 生成测试密钥
	key := KeyFromString("test-key")

	// 创建加密中间件
	encryptionMiddleware := EncryptionMiddleware(key)

	// 准备测试数据
	plainText := []byte("Secret message content")

	// 模拟加密后的数据（作为接收到的消息）
	encryptedData, err := encrypt(key, plainText)
	assert.NoError(t, err)

	// 创建模拟上下文
	logger := log.NewDefaultLogger()
	writer := &MockWriter{}

	// 创建模拟processor
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	ctx := &MockContext{
		msgType:   "test",
		requestID: 1,
		rawData:   encryptedData,
		writer:    writer,
		logger:    logger,
		processor: processor,
	}

	// 创建测试处理器
	handlerCalled := false
	var processedData []byte
	nextHandler := func(ctx core.Context) error {
		handlerCalled = true
		// 验证数据已被解密
		processedData = ctx.RawData()
		return nil
	}

	// 应用加密中间件
	enhancedHandler := encryptionMiddleware(nextHandler)

	// 执行处理器
	err = enhancedHandler(ctx)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, plainText, processedData)
}

// 测试加密写入器的Write方法
func TestEncryptedWriter_Write(t *testing.T) {
	// 生成测试密钥
	key := KeyFromString("test-key")

	// 创建模拟writer和processor
	mockWriter := &MockWriter{}
	logger := log.NewDefaultLogger()
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	// 创建加密写入器
	encryptedWriter := &encryptedWriter{
		writer:    mockWriter,
		key:       key,
		logger:    logger,
		processor: processor,
	}

	// 测试写入消息
	testPayload := "Hello, World!"
	err := encryptedWriter.Write("test", testPayload)
	assert.NoError(t, err)
	assert.Equal(t, "test", mockWriter.lastMsgType)
	assert.NotEmpty(t, mockWriter.lastPayload)

	// 验证数据已被加密（通过尝试解密来验证）
	encryptedData, ok := mockWriter.lastPayload.([]byte)
	assert.True(t, ok)

	decryptedData, err := decrypt(key, encryptedData)
	assert.NoError(t, err)

	var decodedPayload string
	err = serializer.DefaultSerializer.Deserialize(decryptedData, &decodedPayload)
	assert.NoError(t, err)
	assert.Equal(t, testPayload, decodedPayload)
}

// 测试加密写入器的Reply方法
func TestEncryptedWriter_Reply(t *testing.T) {
	// 生成测试密钥
	key := KeyFromString("test-key")

	// 创建模拟writer和processor
	mockWriter := &MockWriter{}
	logger := log.NewDefaultLogger()
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	// 创建加密写入器
	encryptedWriter := &encryptedWriter{
		writer:    mockWriter,
		key:       key,
		logger:    logger,
		processor: processor,
	}

	// 测试回复消息
	requestID := uint64(123)
	testPayload := "Reply message"
	err := encryptedWriter.Reply(requestID, "test_reply", testPayload)
	assert.NoError(t, err)
	assert.Equal(t, "test_reply", mockWriter.lastMsgType)
	assert.NotEmpty(t, mockWriter.lastPayload)
}

// MockContext 是 Context 接口的模拟实现
type MockContext struct {
	msgType   string
	requestID uint64
	rawData   []byte
	writer    core.Writer
	logger    log.Logger
	processor core.Processor
}

func (c *MockContext) Bind(target interface{}) error {
	return serializer.DefaultSerializer.Deserialize(c.rawData, target)
}

func (c *MockContext) MessageType() string {
	return c.msgType
}

func (c *MockContext) RequestID() uint64 {
	return c.requestID
}

func (c *MockContext) IsRequest() bool {
	return c.requestID > 0
}

func (c *MockContext) IsResponse() bool {
	return !c.IsRequest()
}

func (c *MockContext) Connection() transport.Connection {
	return nil
}

func (c *MockContext) RawData() []byte {
	return c.rawData
}

func (c *MockContext) SetRawData(data []byte) {
	c.rawData = data
}

func (c *MockContext) Writer() core.Writer {
	return c.writer
}

func (c *MockContext) SetWriter(writer core.Writer) {
	c.writer = writer
}

func (c *MockContext) Processor() core.Processor {
	return c.processor
}

func (c *MockContext) Reply(payload interface{}) error {
	return c.writer.Reply(c.requestID, c.msgType+"_response", payload)
}

func (c *MockContext) Logger() log.Logger {
	return c.logger
}
