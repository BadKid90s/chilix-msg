package transport

import (
	"crypto/sha1"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

type kcpListener struct {
	*kcp.Listener
}

func (l *kcpListener) Accept() (Connection, error) {
	return l.Listener.Accept()
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
	l, err := kcp.ListenWithOptions(address, t.block, 10, 3)
	if err != nil {
		return nil, err
	}
	return &kcpListener{Listener: l}, nil
}

func (t *kcpTransport) Dial(address string) (Connection, error) {
	return kcp.DialWithOptions(address, t.block, 10, 3)
}
