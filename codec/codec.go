package codec

import (
	"errors"
	"io"
)

// Codec 编解码器接口
type Codec interface {
	Encode(w io.Writer, typeID uint32, payload any, requestID uint64) error
	Decode(r io.Reader) (typeID uint32, payload []byte, requestID uint64, err error)
}

var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrUnsupportedVersion   = errors.New("unsupported protocol version")
)
