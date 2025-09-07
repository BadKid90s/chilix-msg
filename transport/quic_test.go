package transport

import (
	"fmt"
	"testing"
	"time"
)

func TestQUICTransport_Protocol(t *testing.T) {
	transport := NewQUICTransport()
	if transport.Protocol() != "quic" {
		t.Errorf("Expected protocol to be 'quic', got '%s'", transport.Protocol())
	}
}

func TestQUICTransport_BasicConnection(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 15 * time.Second, // QUIC需要更长时间建立连接
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunBasicConnectionTest(config)
	if !result.Success {
		t.Fatalf("Basic connection test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("QUIC basic connection test: %s", result.Message)
}

func TestQUICTransport_LargeData(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 20 * time.Second, // QUIC需要更长时间处理大数据
		DataTimeout:       10 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunLargeDataTest(config)
	if !result.Success {
		t.Fatalf("Large data test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("QUIC large data test: %s", result.Message)
}

func TestQUICTransport_Timeout(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 10 * time.Second,
		DataTimeout:       3 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunTimeoutTest(config)
	if !result.Success {
		t.Fatalf("Timeout test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("QUIC timeout test: %s", result.Message)
}

func TestQUICTransport_InvalidAddress(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 10 * time.Second,
		DataTimeout:       3 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunInvalidAddressTest(t, config)
	if !result.Success {
		t.Fatalf("Invalid address test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("QUIC invalid address test: %s", result.Message)
}

func TestQUICListener_Addr(t *testing.T) {
	transport := NewQUICTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer SafeClose(listener, "quic-listener")

	addr := listener.Addr()
	if addr == nil {
		t.Fatal("Expected non-nil address")
	}

	// QUIC监听器底层使用UDP，所以网络类型应该是udp
	if addr.Network() != "udp" {
		t.Errorf("Expected network to be 'udp', got '%s'", addr.Network())
	}
}

func TestQUICListener_Close(t *testing.T) {
	transport := NewQUICTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// 在协程中关闭监听器，单独处理错误
	SafeClose(listener, "quic-listener")

	// 等待一下让关闭操作完成
	time.Sleep(100 * time.Millisecond)

	// 尝试接受连接，应该失败
	_, err = listener.Accept()
	if err == nil {
		t.Fatal("Expected Accept to fail after listener closed")
	}
}

func TestQUICTransport_ConnectionInterface(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 15 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "quic-echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer SafeClose(clientConn, "quic-client-connection")

	// 测试连接的基本属性
	if clientConn.LocalAddr() == nil {
		t.Errorf("LocalAddr() returned nil")
	}
	if clientConn.RemoteAddr() == nil {
		t.Errorf("RemoteAddr() returned nil")
	}

	// 测试设置超时
	err = clientConn.SetDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetDeadline() failed: %v", err)
	}

	err = clientConn.SetReadDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetReadDeadline() failed: %v", err)
	}

	err = clientConn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		t.Errorf("SetWriteDeadline() failed: %v", err)
	}

	// 重置超时
	err = clientConn.SetDeadline(time.Time{})
	if err != nil {
		t.Errorf("Failed to reset deadline: %v", err)
	}

	t.Log("QUIC connection interface test passed")
}

func TestQUICTransport_TLSConfig(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 15 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "quic-echo-server")

	// 客户端连接 - 这将使用TLS
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial with TLS: %v", err)
	}
	defer SafeClose(clientConn, "quic-client-connection")

	// 设置连接超时
	_ = clientConn.SetDeadline(time.Now().Add(config.ConnectionTimeout))

	// 测试数据传输 - 验证TLS加密通道工作正常
	testData := []byte("Secure QUIC data over TLS!")

	// 发送数据
	_, err = clientConn.Write(testData)
	if err != nil {
		t.Fatalf("Failed to write data over TLS: %v", err)
	}

	// 接收回显数据
	buf := make([]byte, len(testData))
	_, err = clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read data over TLS: %v", err)
	}

	// 验证数据
	if string(buf) != string(testData) {
		t.Errorf("TLS data mismatch. Expected: %s, Got: %s", testData, buf)
	}

	t.Log("QUIC TLS config test passed")
}

func TestQUICTransport_Multiplexing(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "QUIC",
		Transport:         NewQUICTransport(),
		ConnectionTimeout: 15 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "quic-echo-server")

	// 创建多个客户端连接，测试QUIC的多路复用能力
	connections := make([]Connection, 3)
	for i := 0; i < 3; i++ {
		conn, err := config.Transport.Dial(server.Addr().String())
		if err != nil {
			t.Fatalf("Failed to dial server (connection %d): %v", i, err)
		}
		connections[i] = conn
	}

	// 在函数结束时统一关闭所有连接
	defer func() {
		for i, conn := range connections {
			if conn != nil {
				SafeClose(conn, fmt.Sprintf("quic-client-connection-%d", i))
			}
		}
	}()

	// 在所有连接上同时发送数据
	for i, conn := range connections {
		testData := []byte("QUIC multiplexing test " + string(rune('0'+i)))

		// 发送数据
		_, err = conn.Write(testData)
		if err != nil {
			t.Fatalf("Failed to write data (connection %d): %v", i, err)
		}

		// 接收回显数据
		buf := make([]byte, len(testData))
		_, err = conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read data (connection %d): %v", i, err)
		}

		// 验证数据
		if string(buf) != string(testData) {
			t.Errorf("Data mismatch (connection %d): expected %s, got %s", i, testData, buf)
		}
	}

	t.Log("QUIC multiplexing test passed")
}
