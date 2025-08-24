package codec

import (
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/BadKid90s/chilix-msg/pkg/serializer"
)

// 头部结构：
// [0:4] 总长度 (uint32, big-endian)
// [4:6] 消息类型长度 (uint16, big-endian)
// [6:6+N] 消息类型 (UTF-8字符串)
// [6+N:14+N] 请求ID (uint64, big-endian)
// [14+N:] 负载数据

const (
	HeaderBaseSize = 14  // 基础头部大小（不含消息类型）
	MaxTypeLength  = 255 // 最大消息类型长度
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

	// 写入总长度
	binary.BigEndian.PutUint32(header[0:4], totalLength)

	// 写入消息类型长度
	binary.BigEndian.PutUint16(header[4:6], uint16(typeLen))

	// 写入消息类型
	copy(header[6:6+typeLen], typeBytes)

	// 写入请求ID
	binary.BigEndian.PutUint64(header[6+typeLen:6+typeLen+8], requestID)

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

	// 读取基础头部
	baseHeader := make([]byte, 6)
	if _, err := io.ReadFull(r, baseHeader); err != nil {
		return "", nil, 0, err
	}

	// 解析总长度
	totalLength := binary.BigEndian.Uint32(baseHeader[0:4])

	// 解析消息类型长度
	typeLen := binary.BigEndian.Uint16(baseHeader[4:6])

	// 检查消息类型长度
	if typeLen > MaxTypeLength {
		return "", nil, 0, ErrMessageTypeTooLong
	}

	// 创建完整头部缓冲区
	fullHeaderSize := 6 + int(typeLen) + 8
	fullHeader := make([]byte, fullHeaderSize)

	// 复制基础头部
	copy(fullHeader[0:6], baseHeader)

	// 读取剩余头部
	if _, err := io.ReadFull(r, fullHeader[6:fullHeaderSize]); err != nil {
		return "", nil, 0, err
	}

	// 解析消息类型
	msgType := string(fullHeader[6 : 6+typeLen])

	// 验证消息类型
	if !isValidMessageType(msgType) {
		return "", nil, 0, ErrInvalidMessageType
	}

	// 解析请求ID
	requestID := binary.BigEndian.Uint64(fullHeader[6+typeLen : 6+typeLen+8])

	// 计算负载长度
	payloadLength := int(totalLength) - fullHeaderSize
	if payloadLength < 0 {
		return "", nil, 0, ErrInvalidMessageFormat
	}

	// 读取负载
	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return "", nil, 0, err
	}
	return msgType, payload, requestID, nil
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
