package transport

import (
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/BadKid90s/chilix-msg/log"
	"github.com/gorilla/websocket"
)

// wsConn is a wrapper around a gorilla/websocket Conn that implements the net.Conn interface.
type wsConn struct {
	conn   *websocket.Conn
	reader io.Reader
	readMu sync.Mutex
}

func (c *wsConn) Read(b []byte) (n int, err error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	for {
		if c.reader != nil {
			n, err = c.reader.Read(b)
			if err == io.EOF {
				c.reader = nil // Current message is done, get next one
				if n > 0 {
					return n, nil // Return what we have, next Read will get next message
				}
				// if n is 0, we continue to get the next message in this call
			} else {
				return n, err
			}
		}

		// No reader or current reader is exhausted, get the next one.
		mt, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}

		if mt != websocket.BinaryMessage {
			// For simplicity, we only handle binary messages.
			// We could skip other message types.
			continue
		}

		c.reader = r
	}
}

func (c *wsConn) Write(b []byte) (n int, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// wsListener implements the net.Listener interface for WebSocket connections.
type wsListener struct {
	addr     net.Addr
	connChan chan Connection
	server   *http.Server
	once     sync.Once
	closeCh  chan struct{}
	ln       net.Listener
}

func (l *wsListener) Accept() (Connection, error) {
	select {
	case conn := <-l.connChan:
		return conn, nil
	case <-l.closeCh:
		return nil, &net.OpError{Op: "accept", Net: "websocket", Addr: l.addr, Err: net.ErrClosed}
	}
}

func (l *wsListener) Close() error {
	l.once.Do(func() {
		close(l.closeCh)
		if l.server != nil {
			// 在协程中关闭服务器，单独处理错误
			go func() {
				if err := l.server.Close(); err != nil {
					// 记录错误但不阻塞
					log.Errorf("Error closing server: %v", err)
				}
			}()
		} else if l.ln != nil {
			// 在协程中关闭监听器，单独处理错误
			go func() {
				if err := l.ln.Close(); err != nil {
					// 记录错误但不阻塞
					log.Errorf("Error closing listener: %v", err)
				}
			}()
		}
	})
	return nil
}

func (l *wsListener) Addr() net.Addr {
	return l.addr
}

type wsTransport struct{}

func (t *wsTransport) Protocol() string {
	return "websocket"
}

// NewWebSocketTransport creates new transport that uses WebSockets.
func NewWebSocketTransport() Transport {
	return &wsTransport{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func (t *wsTransport) Listen(address string) (Listener, error) {
	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	listener := &wsListener{
		addr:     tcpListener.Addr(),
		connChan: make(chan Connection, 128),
		closeCh:  make(chan struct{}),
		ln:       tcpListener,
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				// upgrader writes http error response
				return
			}
			listener.connChan <- &wsConn{conn: conn}
		}),
	}
	listener.server = server

	go func() {
		if err := server.Serve(tcpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return
		}
	}()

	return listener, nil
}

func (t *wsTransport) Dial(address string) (Connection, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://"+address, nil)
	if err != nil {
		return nil, err
	}
	return &wsConn{conn: conn}, nil
}
