package core

import "errors"

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrHandlerNotFound = errors.New("no handler for message type")
	ErrHandlerPanic    = errors.New("handler panicked")
)
