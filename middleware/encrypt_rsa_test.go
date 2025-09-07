package middleware

import (
	"testing"

	"github.com/BadKid90s/chilix-msg/core"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRSAEncryptionMiddleware 测试RSA加密中间件
func TestRSAEncryptionMiddleware(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)
	require.NotNil(t, privateKey)
	require.NotNil(t, publicKey)

	// 创建RSA加密中间件
	rsaMiddleware := RSAEncryptionMiddleware(privateKey, publicKey)

	// 准备测试数据
	plainText := []byte("Secret message for RSA encryption")

	// 模拟加密后的数据（作为接收到的消息）
	encryptedData, err := rsaEncrypt(publicKey, plainText)
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

	// 应用RSA加密中间件
	enhancedHandler := rsaMiddleware(nextHandler)

	// 执行处理器
	err = enhancedHandler(ctx)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, plainText, processedData)
}

// TestRsaEncryptDecrypt 测试RSA加解密功能
func TestRsaEncryptDecrypt(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 测试数据
	originalData := []byte("This is a test message for RSA encryption")

	// 加密数据
	encryptedData, err := rsaEncrypt(publicKey, originalData)
	assert.NoError(t, err)
	assert.NotEqual(t, originalData, encryptedData)

	// 解密数据
	decryptedData, err := rsaDecrypt(privateKey, encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, originalData, decryptedData)
}

// TestRSAKeyGeneration 测试RSA密钥对生成
func TestRSAKeyGeneration(t *testing.T) {
	// 测试生成不同长度的密钥对
	keySizes := []int{1024, 2048, 4096}

	for _, keySize := range keySizes {
		privateKey, publicKey, err := GenerateRSAKeyPair(keySize)
		assert.NoError(t, err)
		assert.NotNil(t, privateKey)
		assert.NotNil(t, publicKey)
		assert.Equal(t, keySize, privateKey.N.BitLen())
		assert.Equal(t, keySize, publicKey.N.BitLen())
	}
}

// TestRSAKeyExportImport 测试RSA密钥的导出和导入
func TestRSAKeyExportImport(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 导出密钥
	privateKeyPEM := ExportRSAPrivateKey(privateKey)
	publicKeyPEM := ExportRSAPublicKey(publicKey)

	assert.NotEmpty(t, privateKeyPEM)
	assert.NotEmpty(t, publicKeyPEM)

	// 导入密钥
	importedPrivateKey, err := LoadRSAPrivateKey(privateKeyPEM)
	assert.NoError(t, err)
	assert.NotNil(t, importedPrivateKey)

	importedPublicKey, err := LoadRSAPublicKey(publicKeyPEM)
	assert.NoError(t, err)
	assert.NotNil(t, importedPublicKey)

	// 验证导入的密钥与原始密钥相同
	assert.Equal(t, privateKey.N.Cmp(importedPrivateKey.N), 0)
	assert.Equal(t, privateKey.E, importedPrivateKey.E)
	assert.Equal(t, privateKey.D.Cmp(importedPrivateKey.D), 0)

	assert.Equal(t, publicKey.N.Cmp(importedPublicKey.N), 0)
	assert.Equal(t, publicKey.E, importedPublicKey.E)
}

// TestRSAEncryptedWriter 测试RSA加密写入器
func TestRSAEncryptedWriter(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 创建模拟writer和processor
	mockWriter := &MockWriter{}
	logger := log.NewDefaultLogger()
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	// 创建RSA加密写入器
	rsaWriter := &rsaEncryptedWriter{
		writer:    mockWriter,
		publicKey: publicKey,
		logger:    logger,
		processor: processor,
	}

	// 测试写入消息
	testPayload := "Hello, RSA World!"
	err = rsaWriter.Write("test", testPayload)
	assert.NoError(t, err)
	assert.Equal(t, "test", mockWriter.lastMsgType)
	assert.NotEmpty(t, mockWriter.lastPayload)

	// 验证数据已被加密（通过尝试解密来验证）
	encryptedData, ok := mockWriter.lastPayload.([]byte)
	assert.True(t, ok)

	decryptedData, err := rsaDecrypt(privateKey, encryptedData)
	assert.NoError(t, err)

	var decodedPayload string
	err = serializer.DefaultSerializer.Deserialize(decryptedData, &decodedPayload)
	assert.NoError(t, err)
	assert.Equal(t, testPayload, decodedPayload)
}

// TestRSAEncryptedWriterReply 测试RSA加密写入器的Reply方法
func TestRSAEncryptedWriterReply(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 创建模拟writer和processor
	mockWriter := &MockWriter{}
	logger := log.NewDefaultLogger()
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	// 创建RSA加密写入器
	rsaWriter := &rsaEncryptedWriter{
		writer:    mockWriter,
		publicKey: publicKey,
		logger:    logger,
		processor: processor,
	}

	// 测试回复消息
	requestID := uint64(123)
	testPayload := "Reply message"
	err = rsaWriter.Reply(requestID, "test_reply", testPayload)
	assert.NoError(t, err)
	assert.Equal(t, "test_reply", mockWriter.lastMsgType)
	assert.NotEmpty(t, mockWriter.lastPayload)

	// 验证数据已被加密（通过尝试解密来验证）
	encryptedData, ok := mockWriter.lastPayload.([]byte)
	assert.True(t, ok)

	decryptedData, err := rsaDecrypt(privateKey, encryptedData)
	assert.NoError(t, err)

	var decodedPayload string
	err = serializer.DefaultSerializer.Deserialize(decryptedData, &decodedPayload)
	assert.NoError(t, err)
	assert.Equal(t, testPayload, decodedPayload)
}

// TestRSAEncryptedWriterError 测试RSA加密写入器的Error方法
func TestRSAEncryptedWriterError(t *testing.T) {
	// 生成RSA密钥对
	privateKey, publicKey, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 创建模拟writer和processor
	mockWriter := &MockWriter{}
	logger := log.NewDefaultLogger()
	conn := NewMockConnection()
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Logger: logger,
	})

	// 创建RSA加密写入器
	rsaWriter := &rsaEncryptedWriter{
		writer:    mockWriter,
		publicKey: publicKey,
		logger:    logger,
		processor: processor,
	}

	// 测试错误消息
	errorMsg := "Test error message"
	err = rsaWriter.Error(errorMsg)
	assert.NoError(t, err)
	assert.NotEmpty(t, mockWriter.lastPayload)

	// 验证错误消息已被加密（通过尝试解密来验证）
	encryptedData, ok := mockWriter.lastPayload.([]byte)
	assert.True(t, ok)

	decryptedData, err := rsaDecrypt(privateKey, encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, errorMsg, string(decryptedData))
}

// TestRsaDecryptWithInvalidData 测试使用无效数据解密的情况
func TestRsaDecryptWithInvalidData(t *testing.T) {
	// 生成RSA密钥对
	privateKey, _, err := GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// 测试使用太短的数据解密
	_, err = rsaDecrypt(privateKey, []byte("short"))
	assert.Error(t, err)

	// 测试使用格式错误的数据解密
	invalidData := make([]byte, 100)
	_, err = rsaDecrypt(privateKey, invalidData)
	assert.Error(t, err)
}

// TestLoadInvalidRSAPrivateKey 测试加载无效的RSA私钥
func TestLoadInvalidRSAPrivateKey(t *testing.T) {
	// 测试加载nil数据
	_, err := LoadRSAPrivateKey(nil)
	assert.Error(t, err)

	// 测试加载无效PEM数据
	_, err = LoadRSAPrivateKey([]byte("invalid pem data"))
	assert.Error(t, err)
}

// TestLoadInvalidRSAPublicKey 测试加载无效的RSA公钥
func TestLoadInvalidRSAPublicKey(t *testing.T) {
	// 测试加载nil数据
	_, err := LoadRSAPublicKey(nil)
	assert.Error(t, err)

	// 测试加载无效PEM数据
	_, err = LoadRSAPublicKey([]byte("invalid pem data"))
	assert.Error(t, err)
}
