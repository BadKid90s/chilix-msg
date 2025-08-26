package transport

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// quicConn implements the net.Conn interface for a QUIC stream.
type quicConn struct {
	stream     *quic.Stream
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (c *quicConn) Read(p []byte) (n int, err error) {
	return c.stream.Read(p)
}

func (c *quicConn) Write(p []byte) (n int, err error) {
	return c.stream.Write(p)
}

func (c *quicConn) Close() error {
	return c.stream.Close()
}

func (c *quicConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *quicConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *quicConn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *quicConn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

func (c *quicConn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// quicListener implements the transport.Listener interface for QUIC.
type quicListener struct {
	*quic.Listener
}

func (l *quicListener) Accept() (Connection, error) {
	conn, err := l.Listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to accept stream: %w", err)
	}

	return &quicConn{
		stream:     stream,
		localAddr:  conn.LocalAddr(),
		remoteAddr: conn.RemoteAddr(),
	}, nil
}

type quicTransport struct{}

func (t *quicTransport) Protocol() string {
	return "quic"
}

// NewQUICTransport creates a new QUIC transport.
func NewQUICTransport() Transport {
	return &quicTransport{}
}

func (t *quicTransport) Listen(address string) (Listener, error) {
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	listener, err := quic.ListenAddr(address, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	return &quicListener{Listener: listener}, nil
}

func (t *quicTransport) Dial(address string) (Connection, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	conn, err := quic.DialAddr(context.Background(), address, tlsConf, nil)
	if err != nil {
		return nil, err
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &quicConn{
		stream:     stream,
		localAddr:  conn.LocalAddr(),
		remoteAddr: conn.RemoteAddr(),
	}, nil
}

// generateTLSConfig sets up a bare-bones TLS config for the server
func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}
