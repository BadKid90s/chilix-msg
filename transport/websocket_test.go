package transport

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestWebSocketTransport_Basic(t *testing.T) {
	t.Parallel()

	tr := NewWebSocketTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Log("Listener close error:", err)
		}
	}()

	addr := listener.Addr().String()
	fmt.Println("WebSocket listen address:", addr)

	serverDone := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Log("Server goroutine panic:", r)
			}
			close(serverDone)
		}()

		conn, err := listener.Accept()
		if err != nil {
			t.Error("Accept failed:", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Log("Connection close error:", err)
			}
		}()

		buf := make([]byte, 64)
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			t.Error("Read failed:", err)
			return
		}
		if n == 0 {
			t.Error("Read 0 bytes")
			return
		}
		if string(buf[:n]) != "hello websocket" {
			t.Errorf("Expected 'hello websocket', got %q", buf[:n])
		}
	}()

	// WebSocket需要稍长的启动时间（HTTP服务器初始化）
	time.Sleep(100 * time.Millisecond)

	conn, err := tr.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Log("Client connection close error:", err)
		}
	}()

	_, err = conn.Write([]byte("hello websocket"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	select {
	case <-serverDone:
	case <-time.After(3 * time.Second):
		t.Fatal("test timed out")
	}
}

func TestWebSocketTransport_Concurrent(t *testing.T) {
	t.Parallel()

	tr := NewWebSocketTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Log("Listener close error:", err)
		}
	}()

	// 启动服务端处理多个连接
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// 检查是否是监听器关闭导致的错误
				var netErr *net.OpError
				if errors.As(err, &netErr) && errors.Is(netErr.Err, net.ErrClosed) {
					return
				}
				t.Log("Accept error:", err)
				continue
			}

			go func(c Connection) {
				defer func() {
					if err := c.Close(); err != nil {
						t.Log("Connection close error:", err)
					}
				}()

				buf := make([]byte, 64)
				n, _ := c.Read(buf)
				write, err := c.Write(buf[:n])
				if err != nil {
					t.Errorf("Write failed: %v", err)
					return
				}
				if write != n {
					t.Errorf("Write not all bytes: %d != %d", write, n)
				}
			}(conn)
		}
	}()

	// WebSocket需要比TCP稍长的启动时间
	time.Sleep(100 * time.Millisecond)
	addr := listener.Addr().String()

	const numClients = 10
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()

			conn, err := tr.Dial(addr)
			if err != nil {
				t.Errorf("Dial failed: %v", err)
				return
			}
			defer func() {
				if err := conn.Close(); err != nil {
					t.Log("Client connection close error:", err)
				}
			}()

			msg := []byte(fmt.Sprintf("client-%d", id))
			_, err = conn.Write(msg)
			if err != nil {
				t.Errorf("Write failed: %v", err)
				return
			}

			buf := make([]byte, 64)
			n, err := conn.Read(buf)
			if err != nil {
				t.Errorf("Read failed: %v", err)
				return
			}

			if string(buf[:n]) != string(msg) {
				t.Errorf("Expected %q, got %q", msg, buf[:n])
			}
		}(i)
	}

	wg.Wait()
}

func TestWebSocketTransport_Timeout(t *testing.T) {
	t.Parallel()

	tr := NewWebSocketTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Log("Listener close error:", err)
		}
	}()

	t.Run("DialInvalidPort", func(t *testing.T) {
		_, err := tr.Dial("127.0.0.1:99999")
		if err == nil {
			t.Fatal("Expected dial error")
		}
	})

	t.Run("DialInvalidProtocol", func(t *testing.T) {
		// WebSocket需要ws://或wss://协议
		_, err := tr.Dial("http://127.0.0.1:0")
		if err == nil {
			t.Fatal("Expected dial error for invalid protocol")
		}
	})

	t.Run("AcceptAfterClose", func(t *testing.T) {
		err := listener.Close()
		if err != nil {
			t.Fatal("Close failed:", err)
			return
		}
		_, err = listener.Accept()
		if err == nil {
			t.Fatal("Expected accept error after close")
		}
	})
}
