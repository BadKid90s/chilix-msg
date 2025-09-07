// Package codec 提供高性能的消息编解码器实现
//
// BalancedCodec 实现了 CHILIX 平衡协议，这是一个优化的二进制协议格式：
//
// 协议格式 (Balanced Protocol v2):
// ┌─────────────────────────────────────────────────────────────────┐
// │                     Magic Number (32bit)                        │  // "CHPM" (0x4348504D)
// ├─────────────────────────────────────────────────────────────────┤
// │ Version(4bit) │ Flags(4bit)   │        Total Length (24bit)   │  // 版本+标志+总长度
// ├─────────────────────────────────────────────────────────────────┤
// │                        Request ID (64bit)                       │  // 请求ID
// ├─────────────────────────────────────────────────────────────────┤
// │ Type ID (32bit)                                                 │  // 消息类型ID
// ├─────────────────────────────────────────────────────────────────┤
// │ Extension TLV (变长, 可选，如果FlagExtended设置)                 │  // 扩展区
// │  - Type(8bit) + Length(16bit) + Value(变长)                    │
// │  - 可多个TLV，Length=0表示结束                                   │
// ├─────────────────────────────────────────────────────────────────┤
// │                     Payload (变长)                              │  // 消息负载
// └─────────────────────────────────────────────────────────────────┘
//
// 字段说明:
// - Magic Number: 固定值 0x4348504D ("CHPM")，用于协议识别
// - Version: 协议版本号，当前为 2 (4bit)
// - Flags: 标志位 (4bit)
//   - Bit 0: BalancedFlagCompressed (0x1) - 压缩标志
//   - Bit 1: BalancedFlagEncrypted (0x2) - 加密标志
//   - Bit 3: BalancedFlagExtended (0x8) - 扩展区标志
//
// - Total Length: 整个消息的总长度 (24bit, 最大16MB)
// - Request ID: 请求标识符 (64bit)
// - Type ID: 消息类型标识符 (32bit)
// - Extension TLV: 可选的扩展字段，支持多个TLV结构
// - Payload: 实际的消息数据
//
// 特性:
// - 高性能: 固定头部结构，快速解析
// - 可扩展: 支持TLV扩展字段
// - 安全: 支持AES-GCM加密
// - 压缩: 预留压缩标志位
// - 类型优化: 32位类型ID，提升匹配性能
package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/BadKid90s/chilix-msg/serializer"
)

var (
	ErrInvalidMagic     = errors.New("invalid magic number")
	ErrInvalidLength    = errors.New("invalid message length")
	ErrMessageTooLarge  = errors.New("message too large")
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidKey       = errors.New("invalid encryption key")
)

// BalancedCodec 协议常量定义
const (
	// MagicNumber 协议魔数，用于识别 CHILIX 协议消息
	// 值: 0x4348504D，对应 ASCII 字符串 "CHPM" (CHILIX Protocol Message)
	// 位置: 消息头部前4字节
	MagicNumber = 0x4348504D

	// MaxMessageSize 最大消息大小限制 (16MB)
	// 用于防止内存耗尽攻击，确保系统稳定性
	MaxMessageSize = 16 * 1024 * 1024

	// BalancedVersion 平衡协议版本号
	// 当前版本: 2 (4bit)
	// 位置: 第5字节的高4位
	BalancedVersion = 2

	// BalancedHeaderSize 平衡协议基础头部大小
	// 包含: Magic(4) + Version/Flags/Length(4) + RequestID(8) + TypeID(4) = 20字节
	// 注意: 实际头部可能包含扩展TLV，此值为最小头部大小
	BalancedHeaderSize = 21
)

// 平衡协议标志位定义 (4bit)
// 位置: 第5字节的低4位
const (
	// BalancedFlagNone 无特殊标志
	BalancedFlagNone = 0x0

	// BalancedFlagCompressed 压缩标志
	// 当设置时，Payload 数据已压缩
	BalancedFlagCompressed = 0x1

	// BalancedFlagEncrypted 加密标志
	// 当设置时，Payload 数据已使用 AES-GCM 加密
	BalancedFlagEncrypted = 0x2

	// BalancedFlagExtended 扩展区标志
	// 当设置时，消息包含 TLV 扩展字段
	BalancedFlagExtended = 0x8
)

// TLV 扩展字段的 Type-Length-Value 结构
// 用于在消息中携带额外的元数据或配置信息
//
// 格式:
// ┌─────────────┬─────────────┬─────────────┐
// │ Type (8bit) │ Length(16bit)│ Value(变长) │
// └─────────────┴─────────────┴─────────────┘
//
// 字段说明:
// - Type: 扩展字段类型标识符 (8bit)
// - Length: Value 字段的字节长度 (16bit, 大端序)
// - Value: 实际的扩展数据 (Length 字节)
//
// 使用场景:
// - 优先级设置
// - 路由信息
// - 自定义元数据
// - 协议扩展
type TLV struct {
	Type   uint8  // 扩展字段类型 (0-255)
	Length uint16 // 数据长度 (0-65535 字节)
	Value  []byte // 扩展数据内容
}

// Encryptor 加密器接口
type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// AESEncryptor AES加密器实现
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建AES加密器
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &AESEncryptor{key: key}, nil
}

// Encrypt 加密数据
func (e *AESEncryptor) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt 解密数据
func (e *AESEncryptor) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrDecryptionFailed
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// BufferPool 零拷贝辅助
type BufferPool struct {
	pool sync.Pool
	size int
}

func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} { return make([]byte, size) },
		},
		size: size,
	}
}

func (p *BufferPool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *BufferPool) Put(buf []byte) {
	if len(buf) == p.size {
		p.pool.Put(buf)
	}
}

// BalancedCodec 新协议编解码器
type BalancedCodec struct {
	serializer serializer.Serializer
	bufferPool *BufferPool
	encryptor  Encryptor
}

func NewBalancedCodec(serializer serializer.Serializer) *BalancedCodec {
	return &BalancedCodec{
		serializer: serializer,
		bufferPool: NewBufferPool(1024),
		encryptor:  nil,
	}
}

// NewBalancedCodecWithEncryption 创建带加密功能的编解码器
func NewBalancedCodecWithEncryption(serializer serializer.Serializer, encryptor Encryptor) *BalancedCodec {
	return &BalancedCodec{
		serializer: serializer,
		bufferPool: NewBufferPool(1024),
		encryptor:  encryptor,
	}
}

// SetEncryptor 设置加密器
func (c *BalancedCodec) SetEncryptor(encryptor Encryptor) {
	c.encryptor = encryptor
}

// Encode 编码消息
func (c *BalancedCodec) Encode(w io.Writer, typeID uint32, payload interface{}, requestID uint64) error {
	return c.EncodeWithFlags(w, typeID, payload, requestID, BalancedFlagNone, nil)
}

// EncodeWithFlags 带标志位的编码
func (c *BalancedCodec) EncodeWithFlags(w io.Writer, typeID uint32, payload interface{}, requestID uint64, flags uint8, extensions []TLV) error {
	// 步骤1: 序列化负载
	data, err := c.serializer.Serialize(payload)
	if err != nil {
		return err
	}

	// 步骤2: 检查是否需要加密
	if flags&BalancedFlagEncrypted != 0 {
		if c.encryptor == nil {
			return ErrEncryptionFailed
		}
		encryptedData, err := c.encryptor.Encrypt(data)
		if err != nil {
			return err
		}
		data = encryptedData
	}

	// 步骤3: 处理扩展区
	var extData []byte
	extLen := 0
	if len(extensions) > 0 {
		flags |= BalancedFlagExtended
		for _, tlv := range extensions {
			// TLV格式: Type(8bit) + Length(16bit) + Value(变长)
			extData = append(extData, tlv.Type)
			// 使用大端序写入长度
			extData = append(extData, byte(tlv.Length>>8), byte(tlv.Length&0xFF))
			extData = append(extData, tlv.Value...)
			extLen += 3 + len(tlv.Value)
		}
		// 结束标志
		extData = append(extData, 0, 0, 0) // Type=0, Length=0
		extLen += 3
	}

	// 确保 extLen 与 extData 长度一致
	if extLen != len(extData) {
		extLen = len(extData)
	}

	// 步骤4: 计算总长度
	totalLength := BalancedHeaderSize + extLen + len(data)
	if totalLength > MaxMessageSize {
		return ErrMessageTooLarge
	}

	// 步骤5: 构建头部
	header := make([]byte, BalancedHeaderSize)

	// 写入Magic Number
	binary.BigEndian.PutUint32(header[0:4], MagicNumber)

	// 写入版本和标志
	header[4] = (BalancedVersion << 4) | flags

	// 写入总长度（24bit）
	header[5] = byte(totalLength >> 16)
	header[6] = byte(totalLength >> 8)
	header[7] = byte(totalLength)

	// 写入请求ID
	binary.BigEndian.PutUint64(header[8:16], requestID)

	// 写入类型ID
	binary.BigEndian.PutUint32(header[16:20], typeID)

	// 步骤6: 写入数据
	if _, err := w.Write(header); err != nil {
		return err
	}
	// 写入扩展区
	if extLen > 0 {
		if _, err := w.Write(extData); err != nil {
			return err
		}
	}
	// 写入负载
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

// Decode 解码消息
func (c *BalancedCodec) Decode(r io.Reader) (uint32, []byte, uint64, error) {
	typeID, payload, requestID, _, _, err := c.DecodeWithFlags(r)
	return typeID, payload, requestID, err
}

// DecodeWithFlags 带标志位的解码
func (c *BalancedCodec) DecodeWithFlags(r io.Reader) (uint32, []byte, uint64, uint8, []TLV, error) {
	// 步骤1: 读取基础头部
	header := make([]byte, BalancedHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, 0, 0, nil, err
	}

	// 步骤2: 验证Magic Number
	magic := binary.BigEndian.Uint32(header[0:4])
	if magic != MagicNumber {
		return 0, nil, 0, 0, nil, ErrInvalidMagic
	}

	// 步骤3: 解析版本和标志
	versionFlags := header[4]
	version := versionFlags >> 4
	flags := versionFlags & 0x0F
	if version != BalancedVersion {
		return 0, nil, 0, 0, nil, ErrUnsupportedVersion
	}

	// 步骤4: 解析总长度
	totalLength := int(header[5])<<16 | int(header[6])<<8 | int(header[7])
	if totalLength < BalancedHeaderSize || totalLength > MaxMessageSize {
		return 0, nil, 0, 0, nil, ErrInvalidLength
	}

	// 步骤5: 解析请求ID
	requestID := binary.BigEndian.Uint64(header[8:16])

	// 步骤6: 解析类型ID
	typeID := binary.BigEndian.Uint32(header[16:20])

	// 步骤7: 读取扩展区（如果有）
	extLen := 0
	var extensions []TLV
	if flags&BalancedFlagExtended != 0 {
		for {
			tlvHeader := make([]byte, 3) // Type(8bit) + Length(16bit)
			if _, err := io.ReadFull(r, tlvHeader); err != nil {
				return 0, nil, 0, 0, nil, err
			}
			tlvType := tlvHeader[0]
			tlvLen := binary.BigEndian.Uint16(tlvHeader[1:3])
			if tlvLen == 0 {
				break // 结束标志
			}
			value := make([]byte, tlvLen)
			if _, err := io.ReadFull(r, value); err != nil {
				return 0, nil, 0, 0, nil, err
			}
			extLen += 3 + int(tlvLen)
			extensions = append(extensions, TLV{
				Type:   tlvType,
				Length: tlvLen,
				Value:  value,
			})
		}
		extLen += 3 // 结束标志
	}

	// 步骤8: 计算并读取负载
	payloadLength := totalLength - BalancedHeaderSize - extLen
	if payloadLength < 0 {
		return 0, nil, 0, 0, nil, ErrInvalidMessageFormat
	}
	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, 0, 0, nil, err
	}

	// 步骤9: 检查是否需要解密
	if flags&BalancedFlagEncrypted != 0 {
		if c.encryptor == nil {
			return 0, nil, 0, 0, nil, ErrDecryptionFailed
		}
		decryptedPayload, err := c.encryptor.Decrypt(payload)
		if err != nil {
			return 0, nil, 0, 0, nil, err
		}
		payload = decryptedPayload
	}

	return typeID, payload, requestID, flags, extensions, nil
}
