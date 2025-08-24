package transport

import "errors"

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrListenFailed     = errors.New("listen failed")
	ErrAcceptFailed     = errors.New("accept connection failed")
)
