package codec

import (
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/BadKid90s/chilix-msg/serializer"
)

// 优化后的头部结构：
// [0:1] 版本 (uint8)
// [1:2] 标志 (uint8) - 压缩、加密等特性
// [2:6] 总长度 (uint32, big-endian)
// [6:14] 请求ID (uint64, big-endian, 8字节对齐)
// [14:15] 消息类型长度 (uint8)
// [15:15+N] 消息类型 (UTF-8字符串)
// [15+N:] 负载数据

const (
	ProtocolVersion = 1      // 协议版本
	HeaderBaseSize  = 15     // 基础头部大小（不含消息类型）
	MaxTypeLength   = 255    // 最大消息类型长度
)

// 标志位定义
const (
	FlagNone       = 0x00 // 无特殊标志
	FlagCompressed = 0x01 // 压缩标志
	FlagEncrypted  = 0x02 // 加密标志
	// 预留其他标志位...
)

// Codec 编解码器接口
type Codec interface {
	Encode(w io.Writer, msgType string, payload interface{}, requestID uint64) error
	Decode(r io.Reader) (msgType string, payload []byte, requestID uint64, err error)
}

// LengthPrefixCodec 基于长度前缀的编解码器实现
type LengthPrefixCodec struct {
	serializer serializer.Serializer
}

func NewLengthPrefixCodec(serializer serializer.Serializer) *LengthPrefixCodec {
	return &LengthPrefixCodec{
		serializer: serializer,
	}
}

func (c *LengthPrefixCodec) Encode(w io.Writer, msgType string, payload interface{}, requestID uint64) error {
	return c.EncodeWithFlags(w, msgType, payload, requestID, FlagNone)
}

func (c *LengthPrefixCodec) EncodeWithFlags(w io.Writer, msgType string, payload interface{}, requestID uint64, flags uint8) error {
	// 验证消息类型
	if !isValidMessageType(msgType) {
		return ErrInvalidMessageType
	}

	// 序列化payload
	data, err := c.serializer.Serialize(payload)
	if err != nil {
		return err
	}

	// 获取消息类型字节
	typeBytes := []byte(msgType)
	typeLen := len(typeBytes)

	// 检查消息类型长度
	if typeLen > MaxTypeLength {
		return ErrMessageTypeTooLong
	}

	// 计算总长度
	totalLength := uint32(HeaderBaseSize + typeLen + len(data))

	// 创建头部缓冲区
	header := make([]byte, HeaderBaseSize+typeLen)

	// 写入版本
	header[0] = ProtocolVersion

	// 写入标志
	header[1] = flags

	// 写入总长度
	binary.BigEndian.PutUint32(header[2:6], totalLength)

	// 写入请求ID（固定位置，8字节对齐）
	binary.BigEndian.PutUint64(header[6:14], requestID)

	// 写入消息类型长度
	header[14] = uint8(typeLen)

	// 写入消息类型
	copy(header[15:15+typeLen], typeBytes)

	// 写入头部
	if _, err := w.Write(header); err != nil {
		return err
	}

	// 写入负载
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func (c *LengthPrefixCodec) Decode(r io.Reader) (string, []byte, uint64, error) {
	msgType, payload, requestID, _, err := c.DecodeWithFlags(r)
	return msgType, payload, requestID, err
}

func (c *LengthPrefixCodec) DecodeWithFlags(r io.Reader) (string, []byte, uint64, uint8, error) {
	// 读取固定头部
	fixedHeader := make([]byte, 15)
	if _, err := io.ReadFull(r, fixedHeader); err != nil {
		return "", nil, 0, 0, err
	}

	// 检查版本
	version := fixedHeader[0]
	if version != ProtocolVersion {
		return "", nil, 0, 0, ErrUnsupportedVersion
	}

	// 解析标志
	flags := fixedHeader[1]

	// 解析总长度
	totalLength := binary.BigEndian.Uint32(fixedHeader[2:6])

	// 解析请求ID
	requestID := binary.BigEndian.Uint64(fixedHeader[6:14])

	// 解析消息类型长度
	typeLen := fixedHeader[14]

	// 检查消息类型长度
	if typeLen > MaxTypeLength {
		return "", nil, 0, 0, ErrMessageTypeTooLong
	}

	// 读取消息类型
	msgTypeBytes := make([]byte, typeLen)
	if _, err := io.ReadFull(r, msgTypeBytes); err != nil {
		return "", nil, 0, 0, err
	}

	// 解析消息类型
	msgType := string(msgTypeBytes)

	// 验证消息类型
	if !isValidMessageType(msgType) {
		return "", nil, 0, 0, ErrInvalidMessageType
	}

	// 计算负载长度
	headerSize := HeaderBaseSize + int(typeLen)
	payloadLength := int(totalLength) - headerSize
	if payloadLength < 0 {
		return "", nil, 0, 0, ErrInvalidMessageFormat
	}

	// 读取负载
	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return "", nil, 0, 0, err
	}

	return msgType, payload, requestID, flags, nil
}

// 验证消息类型是否有效
func isValidMessageType(msgType string) bool {
	// 检查长度
	if len(msgType) == 0 || len(msgType) > MaxTypeLength {
		return false
	}

	// 检查是否只包含有效字符
	for _, r := range msgType {
		if r < 32 || r > 126 {
			return false
		}
	}

	// 检查是否为有效UTF-8
	if !utf8.ValidString(msgType) {
		return false
	}

	// 检查是否包含控制字符
	if strings.ContainsAny(msgType, "\x00-\x1F\x7F") {
		return false
	}

	return true
}
