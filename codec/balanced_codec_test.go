package codec_test

import (
	"bytes"
	"testing"

	"github.com/BadKid90s/chilix-msg/codec"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalancedCodecRoundTrip(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s)

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 编码消息
	typeID := uint32(12345)
	payload := map[string]string{"key": "value", "number": "123"}
	requestID := uint64(12345)
	err := c.Encode(buf, typeID, payload, requestID)
	require.NoError(t, err)

	// 解码消息
	decodedTypeID, decodedPayload, decodedRequestID, err := c.Decode(buf)
	require.NoError(t, err)

	// 验证结果
	assert.Equal(t, typeID, decodedTypeID)
	assert.Equal(t, requestID, decodedRequestID)

	// 验证负载
	var decodedData map[string]string
	err = s.Deserialize(decodedPayload, &decodedData)
	require.NoError(t, err)
	assert.Equal(t, "value", decodedData["key"])
	assert.Equal(t, "123", decodedData["number"])
}

func TestBalancedCodecWithEncryption(t *testing.T) {
	s := serializer.DefaultSerializer

	// 创建AES加密器
	key := []byte("1234567890123456") // 16字节密钥
	encryptor, err := codec.NewAESEncryptor(key)
	require.NoError(t, err)

	c := codec.NewBalancedCodecWithEncryption(s, encryptor)

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 编码消息（带加密）
	typeID := uint32(67890)
	payload := "sensitive data"
	requestID := uint64(67890)
	flags := uint8(codec.BalancedFlagEncrypted)
	err = c.EncodeWithFlags(buf, typeID, payload, requestID, flags, nil)
	require.NoError(t, err)

	// 解码消息
	decodedTypeID, decodedPayload, decodedRequestID, decodedFlags, extensions, err := c.DecodeWithFlags(buf)
	require.NoError(t, err)

	// 验证结果
	assert.Equal(t, typeID, decodedTypeID)
	assert.Equal(t, requestID, decodedRequestID)
	assert.Equal(t, flags, decodedFlags)
	assert.Empty(t, extensions)

	// 验证负载
	var decodedData string
	err = s.Deserialize(decodedPayload, &decodedData)
	require.NoError(t, err)
	assert.Equal(t, "sensitive data", decodedData)
}

func TestBalancedCodecEncryptionWithoutEncryptor(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s) // 没有加密器

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 尝试编码加密消息（应该失败）
	typeID := uint32(12345)
	payload := "test data"
	requestID := uint64(12345)
	flags := uint8(codec.BalancedFlagEncrypted)
	err := c.EncodeWithFlags(buf, typeID, payload, requestID, flags, nil)
	assert.Error(t, err)
	assert.Equal(t, codec.ErrEncryptionFailed, err)
}

func TestBalancedCodecDecryptionWithoutEncryptor(t *testing.T) {
	s := serializer.DefaultSerializer

	// 创建加密器用于编码
	key := []byte("1234567890123456")
	encryptor, err := codec.NewAESEncryptor(key)
	require.NoError(t, err)

	c1 := codec.NewBalancedCodecWithEncryption(s, encryptor)
	c2 := codec.NewBalancedCodec(s) // 没有加密器

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 用加密器编码消息
	typeID := uint32(12345)
	payload := "test data"
	requestID := uint64(12345)
	flags := uint8(codec.BalancedFlagEncrypted)
	err = c1.EncodeWithFlags(buf, typeID, payload, requestID, flags, nil)
	require.NoError(t, err)

	// 尝试用没有加密器的解码器解码（应该失败）
	_, _, _, _, _, err = c2.DecodeWithFlags(buf)
	assert.Error(t, err)
	assert.Equal(t, codec.ErrDecryptionFailed, err)
}

func TestAESEncryptor(t *testing.T) {
	// 测试不同长度的密钥
	testCases := []struct {
		name  string
		key   []byte
		valid bool
	}{
		{"16字节密钥", []byte("1234567890123456"), true},
		{"24字节密钥", []byte("123456789012345678901234"), true},
		{"32字节密钥", []byte("12345678901234567890123456789012"), true},
		{"15字节密钥", []byte("123456789012345"), false},
		{"17字节密钥", []byte("12345678901234567"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encryptor, err := codec.NewAESEncryptor(tc.key)
			if tc.valid {
				require.NoError(t, err)
				require.NotNil(t, encryptor)

				// 测试加密解密
				originalData := []byte("test data for encryption")
				encryptedData, err := encryptor.Encrypt(originalData)
				require.NoError(t, err)
				require.NotEqual(t, originalData, encryptedData)

				decryptedData, err := encryptor.Decrypt(encryptedData)
				require.NoError(t, err)
				assert.Equal(t, originalData, decryptedData)
			} else {
				assert.Error(t, err)
				assert.Equal(t, codec.ErrInvalidKey, err)
			}
		})
	}
}

func TestBalancedCodecWithExtensions(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s)

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 创建扩展数据
	extensions := []codec.TLV{
		{Type: 1, Length: 4, Value: []byte("test")},
		{Type: 2, Length: 3, Value: []byte("ext")},
	}

	// 编码消息带扩展
	typeID := uint32(11111)
	payload := map[string]int{"count": 42}
	requestID := uint64(11111)
	flags := uint8(codec.BalancedFlagExtended)
	err := c.EncodeWithFlags(buf, typeID, payload, requestID, flags, extensions)
	require.NoError(t, err)

	// 解码消息
	decodedTypeID, decodedPayload, decodedRequestID, decodedFlags, decodedExtensions, err := c.DecodeWithFlags(buf)
	require.NoError(t, err)

	// 验证结果
	assert.Equal(t, typeID, decodedTypeID)
	assert.Equal(t, requestID, decodedRequestID)
	assert.Equal(t, flags, decodedFlags)
	assert.Len(t, decodedExtensions, 2)

	// 验证扩展数据
	assert.Equal(t, uint8(1), decodedExtensions[0].Type)
	assert.Equal(t, uint16(4), decodedExtensions[0].Length)
	assert.Equal(t, []byte("test"), decodedExtensions[0].Value)

	assert.Equal(t, uint8(2), decodedExtensions[1].Type)
	assert.Equal(t, uint16(3), decodedExtensions[1].Length)
	assert.Equal(t, []byte("ext"), decodedExtensions[1].Value)

	// 验证负载
	var decodedData map[string]int
	err = s.Deserialize(decodedPayload, &decodedData)
	require.NoError(t, err)
	assert.Equal(t, 42, decodedData["count"])
}

func TestBalancedCodecInvalidMagic(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s)

	// 创建无效的头部数据
	buf := &bytes.Buffer{}
	invalidHeader := make([]byte, codec.BalancedHeaderSize)
	// 写入错误的Magic Number
	invalidHeader[0] = 0xFF
	invalidHeader[1] = 0xFF
	invalidHeader[2] = 0xFF
	invalidHeader[3] = 0xFF
	buf.Write(invalidHeader)

	// 尝试解码
	_, _, _, err := c.Decode(buf)
	assert.Error(t, err)
	assert.Equal(t, codec.ErrInvalidMagic, err)
}

func TestBalancedCodecUnsupportedVersion(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s)

	// 创建无效版本的数据
	buf := &bytes.Buffer{}
	invalidHeader := make([]byte, codec.BalancedHeaderSize)
	// 写入正确的Magic Number
	invalidHeader[0] = 0x43 // 'C'
	invalidHeader[1] = 0x48 // 'H'
	invalidHeader[2] = 0x50 // 'P'
	invalidHeader[3] = 0x4D // 'M'
	// 写入错误的版本号
	invalidHeader[4] = 0x10 // 版本1，但应该是版本2
	buf.Write(invalidHeader)

	// 尝试解码
	_, _, _, err := c.Decode(buf)
	assert.Error(t, err)
	assert.Equal(t, codec.ErrUnsupportedVersion, err)
}

func TestBalancedCodecBufferPool(t *testing.T) {
	// 测试BufferPool
	pool := codec.NewBufferPool(1024)

	// 获取缓冲区
	buf1 := pool.Get()
	assert.Len(t, buf1, 1024)

	// 归还缓冲区
	pool.Put(buf1)

	// 再次获取缓冲区
	buf2 := pool.Get()
	assert.Len(t, buf2, 1024)

	// 归还缓冲区
	pool.Put(buf2)
}

func TestBalancedCodecMultipleMessages(t *testing.T) {
	s := serializer.DefaultSerializer
	c := codec.NewBalancedCodec(s)

	// 创建缓冲区
	buf := &bytes.Buffer{}

	// 编码多个消息
	messages := []struct {
		typeID    uint32
		payload   interface{}
		requestID uint64
	}{
		{1, "payload1", 1},
		{2, map[string]int{"id": 2}, 2},
		{3, []string{"a", "b", "c"}, 3},
	}

	for _, msg := range messages {
		err := c.Encode(buf, msg.typeID, msg.payload, msg.requestID)
		require.NoError(t, err)
	}

	// 解码所有消息
	for i, expected := range messages {
		decodedTypeID, decodedPayload, decodedRequestID, err := c.Decode(buf)
		require.NoError(t, err)

		assert.Equal(t, expected.typeID, decodedTypeID)
		assert.Equal(t, expected.requestID, decodedRequestID)

		// 根据消息类型验证负载
		switch i {
		case 0:
			var decodedData string
			err = s.Deserialize(decodedPayload, &decodedData)
			require.NoError(t, err)
			assert.Equal(t, "payload1", decodedData)
		case 1:
			var decodedData map[string]int
			err = s.Deserialize(decodedPayload, &decodedData)
			require.NoError(t, err)
			assert.Equal(t, 2, decodedData["id"])
		case 2:
			var decodedData []string
			err = s.Deserialize(decodedPayload, &decodedData)
			require.NoError(t, err)
			assert.Equal(t, []string{"a", "b", "c"}, decodedData)
		}
	}
}
