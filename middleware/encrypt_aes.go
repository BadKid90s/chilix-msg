package middleware

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"github.com/BadKid90s/chilix-msg/core"
	"github.com/BadKid90s/chilix-msg/log"
)

// EncryptionMiddleware 加密中间件
func EncryptionMiddleware(key []byte) core.Middleware {
	// 验证密钥长度
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		// 如果密钥长度不正确，生成一个正确的密钥
		key = generateKey(key)
	}

	return func(next core.Handler) core.Handler {
		return func(ctx core.Context) error {
			// 解密接收到的消息
			decryptedData, err := decrypt(key, ctx.RawData())
			if err != nil {
				ctx.Logger().Errorf("Decryption failed: %v", err)
				// 解密失败是框架层错误，直接返回error而不是业务响应
				return err
			}

			// 更新上下文中的原始数据
			ctx.SetRawData(decryptedData)

			// 创建加密写入器
			encryptedWriter := &encryptedWriter{
				writer:    ctx.Writer(),
				key:       key,
				logger:    ctx.Logger(),
				processor: ctx.Processor(),
			}
			ctx.SetWriter(encryptedWriter)

			// 调用下一个处理器
			return next(ctx)
		}
	}
}

// encryptedWriter 加密写入器
type encryptedWriter struct {
	writer    core.Writer
	key       []byte
	logger    log.Logger
	processor *core.Processor
}

func (w *encryptedWriter) Write(msgType string, payload interface{}) error {
	// 使用Processor的序列化器序列化payload
	data, err := w.processor.Serializer().Serialize(payload)
	if err != nil {
		w.logger.Errorf("Failed to serialize payload: %v", err)
		return err
	}

	// 加密数据
	encryptedData, err := encrypt(w.key, data)
	if err != nil {
		w.logger.Errorf("Failed to encrypt data: %v", err)
		return err
	}

	// 发送加密后的数据
	return w.writer.Write(msgType, encryptedData)
}

func (w *encryptedWriter) Reply(requestID uint64, msgType string, payload interface{}) error {
	// 使用Processor的序列化器序列化payload
	data, err := w.processor.Serializer().Serialize(payload)
	if err != nil {
		w.logger.Errorf("Failed to serialize payload: %v", err)
		return err
	}

	// 加密数据
	encryptedData, err := encrypt(w.key, data)
	if err != nil {
		w.logger.Errorf("Failed to encrypt data: %v", err)
		return err
	}

	// 发送加密后的回复
	return w.writer.Reply(requestID, msgType, encryptedData)
}

func (w *encryptedWriter) Error(errorMsg string) error {
	// 加密错误消息
	encryptedError, err := encrypt(w.key, []byte(errorMsg))
	if err != nil {
		w.logger.Errorf("Failed to encrypt error message: %v", err)
		return err
	}

	// 发送加密后的错误消息
	return w.writer.Write("error", encryptedError)
}

// encrypt 加密函数
func encrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt 解密函数
func decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// generateKey 从任意长度的字节生成固定长度的密钥
func generateKey(key []byte) []byte {
	hash := sha256.Sum256(key)
	return hash[:]
}

// KeyFromString 从字符串生成密钥
func KeyFromString(keyString string) []byte {
	return generateKey([]byte(keyString))
}

// KeyFromBase64 从base64字符串生成密钥
func KeyFromBase64(keyString string) ([]byte, error) {
	// 使用base64解码获取密钥
	decoded, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
