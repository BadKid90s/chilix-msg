package transport

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestQUICTransport_Basic(t *testing.T) {
	t.Parallel()

	tr := NewQUICTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func(listener Listener) {
		err := listener.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}(listener)

	addr := listener.Addr().String()
	fmt.Println("QUIC listen address:", addr)

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)

		conn, err := listener.Accept()
		if err != nil {
			t.Error("Accept failed:", err)
			return
		}
		defer func(conn Connection) {
			err := conn.Close()
			if err != nil {
				t.Errorf("Close failed: %v", err)
				return
			}
		}(conn)

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
		if string(buf[:n]) != "hello quic" {
			t.Errorf("Expected 'hello quic', got %q", buf[:n])
		}
	}()

	time.Sleep(200 * time.Millisecond)

	conn, err := tr.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer func(conn Connection) {
		err := conn.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
			return
		}
	}(conn)

	_, err = conn.Write([]byte("hello quic"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	select {
	case <-serverDone:
	case <-time.After(3 * time.Second):
		t.Fatal("test timed out")
	}
}

func TestQUICTransport_Concurrent(t *testing.T) {
	t.Parallel()

	tr := NewQUICTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func(listener Listener) {
		err := listener.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}(listener)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func(c net.Conn) {
					err := c.Close()
					if err != nil {
						t.Errorf("Close failed: %v", err)
						return
					}
				}(c)
				buf := make([]byte, 64)
				n, _ := c.Read(buf)
				write, err := c.Write(buf[:n])
				if err != nil {
					t.Errorf("Write failed: %v", err)
					return
				}
				if write != n {
					t.Errorf("Expected to write %d bytes, but wrote %d", n, write)
					return
				}
			}(conn)
		}
	}()

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
			defer func(conn Connection) {
				err := conn.Close()
				if err != nil {
					t.Errorf("Close failed: %v", err)
					return
				}
			}(conn)

			msg := []byte(fmt.Sprintf("client-%d", id))
			_, err = conn.Write(msg)
			if err != nil {
				t.Errorf("Write failed: %v", err)
				return
			}

			buf := make([]byte, 64)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
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

func TestQUICTransport_Timeout(t *testing.T) {
	t.Parallel()

	tr := NewQUICTransport()
	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer func(listener Listener) {
		err := listener.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}(listener)

	t.Run("DialInvalidPort", func(t *testing.T) {
		_, err := tr.Dial("127.0.0.1:99999")
		if err == nil {
			t.Fatal("Expected dial error")
		}
	})

	t.Run("AcceptAfterClose", func(t *testing.T) {
		err := listener.Close()
		if err != nil {
			t.Errorf("Accept after close")
			return
		}
		_, err = listener.Accept()
		if err == nil {
			t.Fatal("Expected accept error after close")
		}
	})
}
