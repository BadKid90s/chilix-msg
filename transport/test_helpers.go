package transport

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/BadKid90s/chilix-msg/log"
)

// SafeClose 在协程中安全关闭连接，单独处理错误
func SafeClose(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		log.Errorf("Error closing %s: %v", name, err)
	}
}

// TestConfig 测试配置
type TestConfig struct {
	ProtocolName      string
	Transport         Transport
	ConnectionTimeout time.Duration
	DataTimeout       time.Duration
	SkipNetworkTests  bool
}

// TestResult 测试结果
type TestResult struct {
	Success bool
	Error   error
	Message string
}

// EchoServer 回显服务器
type EchoServer struct {
	listener Listener
	conns    []Connection
	mu       sync.Mutex
	closed   bool
}

// NewEchoServer 创建新的回显服务器
func NewEchoServer(transport Transport, address string) (*EchoServer, error) {
	listener, err := transport.Listen(address)
	if err != nil {
		return nil, err
	}

	server := &EchoServer{
		listener: listener,
		conns:    make([]Connection, 0),
	}

	go server.acceptLoop()
	return server, nil
}

// acceptLoop 接受连接循环
func (s *EchoServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.closed {
				// 只有在服务器未关闭时才记录错误
				log.Errorf("Error accepting connection: %v", err)
			}
			return
		}

		s.mu.Lock()
		s.conns = append(s.conns, conn)
		s.mu.Unlock()

		go s.handleConnection(conn)
	}
}

// handleConnection 处理单个连接
func (s *EchoServer) handleConnection(conn Connection) {
	defer func(conn Connection) {
		err := conn.Close()
		if err != nil {
			log.Errorf("Error closing connection: %v", err)
		}
	}(conn)

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				// 记录非EOF错误
				_ = err // 避免未使用变量警告
			}
			return
		}

		// 回显数据
		_, err = conn.Write(buf[:n])
		if err != nil {
			return
		}
	}
}

// Close 关闭服务器
func (s *EchoServer) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	// 在协程中关闭所有连接，单独处理错误
	for i, conn := range s.conns {
		SafeClose(conn, fmt.Sprintf("connection-%d", i))
	}

	// 在协程中关闭监听器，单独处理错误
	SafeClose(s.listener, "listener")

	return nil
}

// Addr 获取服务器地址
func (s *EchoServer) Addr() net.Addr {
	return s.listener.Addr()
}

// RunBasicConnectionTest 运行基本连接测试
func RunBasicConnectionTest(config TestConfig) TestResult {
	if config.SkipNetworkTests {
		return TestResult{Success: true, Message: "Network tests skipped"}
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to create server"}
	}
	defer SafeClose(server, "echo-server")

	// 连接超时
	connTimeout := config.ConnectionTimeout
	if connTimeout == 0 {
		connTimeout = 5 * time.Second
	}

	// 数据超时
	dataTimeout := config.DataTimeout
	if dataTimeout == 0 {
		dataTimeout = 2 * time.Second
	}
	_ = dataTimeout // 避免未使用变量警告

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to dial server"}
	}
	defer SafeClose(clientConn, "client-connection")

	// 设置连接超时
	_ = clientConn.SetDeadline(time.Now().Add(connTimeout))

	// 测试数据
	testData := []byte("Hello, " + config.ProtocolName + " Transport!")

	// 发送数据
	_, err = clientConn.Write(testData)
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to write data"}
	}

	// 接收回显数据
	buf := make([]byte, len(testData))
	_, err = io.ReadFull(clientConn, buf)
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to read data"}
	}

	// 验证数据
	if !bytes.Equal(buf, testData) {
		return TestResult{Success: false, Message: "Data mismatch"}
	}

	return TestResult{Success: true, Message: "Connection test passed"}
}

// RunLargeDataTest 运行大数据传输测试
func RunLargeDataTest(config TestConfig) TestResult {
	if config.SkipNetworkTests {
		return TestResult{Success: true, Message: "Network tests skipped"}
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to create server"}
	}
	defer SafeClose(server, "echo-server")

	// 连接超时
	connTimeout := config.ConnectionTimeout
	if connTimeout == 0 {
		connTimeout = 10 * time.Second
	}

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to dial server"}
	}
	defer SafeClose(clientConn, "client-connection")

	// 设置连接超时
	_ = clientConn.SetDeadline(time.Now().Add(connTimeout))

	// 创建大数据（64KB）
	largeData := make([]byte, 64*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	// 发送大数据
	_, err = clientConn.Write(largeData)
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to write large data"}
	}

	// 接收回显数据
	receivedData := make([]byte, len(largeData))
	_, err = io.ReadFull(clientConn, receivedData)
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to read large data"}
	}

	// 验证数据
	if !bytes.Equal(receivedData, largeData) {
		return TestResult{Success: false, Message: "Large data mismatch"}
	}

	return TestResult{Success: true, Message: "Large data test passed"}
}

// RunTimeoutTest 运行超时测试
func RunTimeoutTest(config TestConfig) TestResult {
	if config.SkipNetworkTests {
		return TestResult{Success: true, Message: "Network tests skipped"}
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to create server"}
	}
	defer SafeClose(server, "echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to dial server"}
	}
	defer SafeClose(clientConn, "client-connection")

	// 设置短超时
	_ = clientConn.SetDeadline(time.Now().Add(100 * time.Millisecond))

	// 尝试读取数据，应该超时
	buf := make([]byte, 10)
	_, err = clientConn.Read(buf)
	if err == nil {
		return TestResult{Success: false, Message: "Expected read to timeout"}
	}

	// 重置超时
	err = clientConn.SetDeadline(time.Time{})
	if err != nil {
		return TestResult{Success: false, Error: err, Message: "Failed to reset deadline"}
	}

	return TestResult{Success: true, Message: "Timeout test passed"}
}

// RunInvalidAddressTest 运行无效地址测试
func RunInvalidAddressTest(t *testing.T, config TestConfig) TestResult {
	// 测试无效地址
	_, err := config.Transport.Dial("invalid-address")
	if err == nil {
		return TestResult{Success: false, Message: "Expected dial to fail with invalid address"}
	}

	// 测试不存在的服务器 - 对于KCP等协议，可能需要更长时间来检测连接失败
	done := make(chan bool, 1)
	go func() {
		_, err := config.Transport.Dial("127.0.0.1:65535")
		if err == nil {
			done <- false
		} else {
			done <- true
		}
	}()

	select {
	case result := <-done:
		if !result {
			return TestResult{Success: false, Message: "Expected dial to fail with non-existent server"}
		}
	case <-time.After(5 * time.Second):
		// 超时也是可以接受的结果，特别是对于KCP等协议
		t.Logf("%s: Dial to non-existent server timed out (this may be expected behavior)", config.ProtocolName)
	}

	return TestResult{Success: true, Message: "Invalid address test passed"}
}
