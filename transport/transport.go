package transport

import (
	"net"
)

// Connection 传输连接接口
type Connection interface {
	net.Conn
}

// Listener 传输监听器接口
type Listener interface {
	Accept() (Connection, error)
	Close() error
	Addr() net.Addr
}

// Transport 传输层接口
type Transport interface {
	Listen(address string) (Listener, error)
	Dial(address string) (Connection, error)
}
