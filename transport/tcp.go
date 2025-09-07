package transport

import "net"

// tcpListener wraps a net.Listener to implement the transport.Listener interface.
type tcpListener struct {
	net.Listener
}

// Accept accepts a new connection and returns it as transport.Connection.
func (l *tcpListener) Accept() (Connection, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *tcpListener) Addr() net.Addr {
	return l.Listener.Addr()
}

type tcpTransport struct{}

func (t *tcpTransport) Protocol() string {
	return "tcp"
}

// NewTCPTransport creates a new TCP transport.
func NewTCPTransport() Transport {
	return &tcpTransport{}
}

func (t *tcpTransport) Listen(address string) (Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpListener{Listener: l}, nil
}

func (t *tcpTransport) Dial(address string) (Connection, error) {
	return net.Dial("tcp", address)
}
