package core

import "errors"

var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrMessageTooLarge  = errors.New("message size exceeds limit")
	ErrRequestTimeout   = errors.New("request timeout")
	ErrHandlerNotFound  = errors.New("no handler for message type")
	ErrInvalidPayload   = errors.New("invalid payload")
)

var (
	ErrHandlerPanic = errors.New("handler panicked")
)
