package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/BadKid90s/chilix-msg/pkg/core"
	"github.com/BadKid90s/chilix-msg/pkg/log"
)

// RSAEncryptionMiddleware 非对称加密中间件
func RSAEncryptionMiddleware(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(ctx core.Context) error {
			// 解密接收到的消息（使用私钥解密）
			decryptedData, err := rsaDecrypt(privateKey, ctx.RawData())
			if err != nil {
				ctx.Logger().Errorf("RSA decryption failed: %v", err)
				return ctx.Error("rsa decryption failed")
			}

			// 更新上下文中的原始数据
			ctx.SetRawData(decryptedData)

			// 创建RSA加密写入器
			rsaWriter := &rsaEncryptedWriter{
				writer:    ctx.Writer(),
				publicKey: publicKey,
				logger:    ctx.Logger(),
				processor: ctx.Processor(),
			}
			ctx.SetWriter(rsaWriter)

			// 调用下一个处理器
			return next(ctx)
		}
	}
}

// rsaEncryptedWriter RSA加密写入器
type rsaEncryptedWriter struct {
	writer    core.Writer
	publicKey *rsa.PublicKey
	logger    log.Logger
	processor *core.Processor
}

func (w *rsaEncryptedWriter) Write(msgType string, payload interface{}) error {
	// 使用Processor的序列化器序列化payload
	data, err := w.processor.Serializer().Serialize(payload)
	if err != nil {
		w.logger.Errorf("Failed to serialize payload: %v", err)
		return err
	}

	// RSA加密数据
	encryptedData, err := rsaEncrypt(w.publicKey, data)
	if err != nil {
		w.logger.Errorf("Failed to encrypt data with RSA: %v", err)
		return err
	}

	// 发送加密后的数据
	return w.writer.Write(msgType, encryptedData)
}

func (w *rsaEncryptedWriter) Reply(requestID uint64, msgType string, payload interface{}) error {
	// 使用Processor的序列化器序列化payload
	data, err := w.processor.Serializer().Serialize(payload)

	if err != nil {
		w.logger.Errorf("Failed to serialize payload: %v", err)
		return err
	}

	// RSA加密数据
	encryptedData, err := rsaEncrypt(w.publicKey, data)
	if err != nil {
		w.logger.Errorf("Failed to encrypt data with RSA: %v", err)
		return err
	}

	// 发送加密后的回复
	return w.writer.Reply(requestID, msgType, encryptedData)
}

func (w *rsaEncryptedWriter) Error(errorMsg string) error {
	// RSA加密错误消息
	encryptedError, err := rsaEncrypt(w.publicKey, []byte(errorMsg))
	if err != nil {
		w.logger.Errorf("Failed to encrypt error message with RSA: %v", err)
		return err
	}

	// 发送加密后的错误消息
	return w.writer.Write("error", encryptedError)
}

// rsaEncrypt RSA加密函数
func rsaEncrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	// RSA加密有长度限制，通常为密钥长度-11字节（使用PKCS#1 v1.5填充时）
	// 因此我们使用混合加密：用RSA加密一个随机AES密钥，然后用AES加密实际数据

	// 生成随机AES密钥
	aesKey := make([]byte, 32) // AES-256
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}

	// 使用AES密钥加密数据
	encryptedData, err := encrypt(aesKey, data)
	if err != nil {
		return nil, err
	}

	// 使用RSA加密AES密钥
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, err
	}

	// 将加密的密钥和加密的数据组合在一起
	// 格式：[密钥长度(4字节)][加密的密钥][加密的数据]
	result := make([]byte, 4+len(encryptedKey)+len(encryptedData))
	// 写入密钥长度
	result[0] = byte(len(encryptedKey) >> 24)
	result[1] = byte(len(encryptedKey) >> 16)
	result[2] = byte(len(encryptedKey) >> 8)
	result[3] = byte(len(encryptedKey))
	// 写入加密的密钥
	copy(result[4:4+len(encryptedKey)], encryptedKey)
	// 写入加密的数据
	copy(result[4+len(encryptedKey):], encryptedData)

	return result, nil
}

// rsaDecrypt RSA解密函数
func rsaDecrypt(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid data format")
	}

	// 读取密钥长度
	keyLen := int(data[0])<<24 | int(data[1])<<16 | int(data[2])<<8 | int(data[3])

	if len(data) < 4+keyLen {
		return nil, errors.New("invalid data format")
	}

	// 读取加密的密钥
	encryptedKey := data[4 : 4+keyLen]

	// 读取加密的数据
	encryptedData := data[4+keyLen:]

	// 使用RSA解密AES密钥
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKey, nil)
	if err != nil {
		return nil, err
	}

	// 使用AES密钥解密数据
	decryptedData, err := decrypt(aesKey, encryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

// LoadRSAPrivateKey 从PEM格式加载RSA私钥
func LoadRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试PKCS#8格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}

		return rsaKey, nil
	}

	return privateKey, nil
}

// LoadRSAPublicKey 从PEM格式加载RSA公钥
func LoadRSAPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		// 尝试PKIX格式
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}

		return rsaKey, nil
	}

	return publicKey, nil
}

// ExportRSAPrivateKey 导出RSA私钥为PEM格式
func ExportRSAPrivateKey(privateKey *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM
}

// ExportRSAPublicKey 导出RSA公钥为PEM格式
func ExportRSAPublicKey(publicKey *rsa.PublicKey) []byte {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM
}
