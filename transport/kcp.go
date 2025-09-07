package transport

import (
	"crypto/sha1"
	"net"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

type kcpListener struct {
	*kcp.Listener
}

func (l *kcpListener) Accept() (Connection, error) {
	conn, err := l.Listener.AcceptKCP()
	if err != nil {
		return nil, err
	}
	// 为接受的连接设置优化参数
	conn.SetNoDelay(1, 10, 2, 1)
	return conn, nil
}

func (l *kcpListener) Addr() net.Addr {
	return l.Listener.Addr()
}

type kcpTransport struct {
	block kcp.BlockCrypt
}

func (t *kcpTransport) Protocol() string {
	return "kcp"
}

// NewKCPTransport creates a new KCP transport.
func NewKCPTransport() Transport {
	key := pbkdf2.Key([]byte("kcp transport pass"), []byte("kcp transport salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	return &kcpTransport{block: block}
}

func (t *kcpTransport) Listen(address string) (Listener, error) {
	l, err := kcp.ListenWithOptions(address, t.block, 0, 0)
	if err != nil {
		return nil, err
	}
	return &kcpListener{Listener: l}, nil
}

func (t *kcpTransport) Dial(address string) (Connection, error) {
	conn, err := kcp.DialWithOptions(address, t.block, 0, 0)
	if err != nil {
		return nil, err
	}
	// 使用与监听端相同的优化参数
	conn.SetNoDelay(1, 10, 2, 1)
	return conn, nil
}
