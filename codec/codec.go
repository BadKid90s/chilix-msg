// Package codec 定义了 CHILIX 消息编解码器的核心接口和错误类型
//
// 该包提供了统一的编解码器接口，支持多种协议实现：
// - BalancedCodec: 平衡协议，支持加密、压缩、扩展字段
// - LengthPrefixCodec: 长度前缀协议，简单高效的固定头部格式
//
// 所有编解码器都遵循相同的接口规范，确保协议间的互操作性。
package codec

import (
	"errors"
	"io"
)

// Codec 编解码器接口
// 定义了 CHILIX 消息编解码器的标准接口
//
// 接口规范:
// - 使用32位类型ID (typeID) 进行消息类型标识
// - 使用64位请求ID (requestID) 进行请求跟踪
// - 支持任意类型的Payload数据
// - 返回原始字节数据，由调用方负责反序列化
//
// 实现要求:
// - 线程安全: 支持并发调用
// - 错误处理: 提供详细的错误信息
// - 性能优化: 高效的内存使用和CPU消耗
type Codec interface {
	// Encode 编码消息到输出流
	// 将消息按照协议格式编码并写入输出流
	//
	// 参数:
	//   - w: 输出流 (io.Writer)
	//   - typeID: 消息类型ID (32bit)
	//   - payload: 消息负载数据 (任意类型)
	//   - requestID: 请求标识符 (64bit)
	// 返回:
	//   - error: 编码过程中的错误
	Encode(w io.Writer, typeID uint32, payload any, requestID uint64) error

	// Decode 从输入流解码消息
	// 从输入流读取并解析消息，返回消息的基本信息
	//
	// 参数:
	//   - r: 输入流 (io.Reader)
	// 返回:
	//   - typeID: 消息类型ID (32bit)
	//   - payload: 原始Payload数据 ([]byte)
	//   - requestID: 请求标识符 (64bit)
	//   - err: 解码过程中的错误
	Decode(r io.Reader) (typeID uint32, payload []byte, requestID uint64, err error)
}

// 编解码器通用错误定义
var (
	// ErrInvalidMessageFormat 消息格式无效错误
	// 当消息格式不符合协议规范时返回
	ErrInvalidMessageFormat = errors.New("invalid message format")

	// ErrUnsupportedVersion 不支持的协议版本错误
	// 当协议版本不被当前实现支持时返回
	ErrUnsupportedVersion = errors.New("unsupported protocol version")
)
