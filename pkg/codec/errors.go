package codec

import "errors"

var (
	ErrMessageTypeTooLong   = errors.New("message type too long")
	ErrInvalidMessageType   = errors.New("invalid message type")
	ErrInvalidMessageFormat = errors.New("invalid message format")
)
