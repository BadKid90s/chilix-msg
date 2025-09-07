package transport

import (
	"net"
	"testing"
	"time"
)

func TestTCPTransport_Protocol(t *testing.T) {
	transport := NewTCPTransport()
	if transport.Protocol() != "tcp" {
		t.Errorf("Expected protocol to be 'tcp', got '%s'", transport.Protocol())
	}
}

func TestTCPTransport_BasicConnection(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "TCP",
		Transport:         NewTCPTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunBasicConnectionTest(config)
	if !result.Success {
		t.Fatalf("Basic connection test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("TCP basic connection test: %s", result.Message)
}

func TestTCPTransport_LargeData(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "TCP",
		Transport:         NewTCPTransport(),
		ConnectionTimeout: 10 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunLargeDataTest(config)
	if !result.Success {
		t.Fatalf("Large data test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("TCP large data test: %s", result.Message)
}

func TestTCPTransport_Timeout(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "TCP",
		Transport:         NewTCPTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunTimeoutTest(config)
	if !result.Success {
		t.Fatalf("Timeout test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("TCP timeout test: %s", result.Message)
}

func TestTCPTransport_InvalidAddress(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "TCP",
		Transport:         NewTCPTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunInvalidAddressTest(t, config)
	if !result.Success {
		t.Fatalf("Invalid address test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("TCP invalid address test: %s", result.Message)
}

func TestTCPListener_Addr(t *testing.T) {
	transport := NewTCPTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer SafeClose(listener, "tcp-listener")

	addr := listener.Addr()
	if addr == nil {
		t.Fatal("Expected non-nil address")
	}

	if addr.Network() != "tcp" {
		t.Errorf("Expected network to be 'tcp', got '%s'", addr.Network())
	}

	// 检查地址格式是否正确
	_, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("Failed to split host:port: %v", err)
	}

	if port == "0" {
		t.Error("Expected non-zero port")
	}
}

func TestTCPListener_Close(t *testing.T) {
	transport := NewTCPTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// 获取地址以便后续测试
	addr := listener.Addr().String()

	// 在协程中关闭监听器，单独处理错误
	SafeClose(listener, "tcp-listener")

	// 等待一下让关闭操作完成
	time.Sleep(100 * time.Millisecond)

	// 尝试连接到已关闭的监听器，应该失败
	_, err = transport.Dial(addr)
	if err == nil {
		t.Fatal("Expected dial to fail after listener closed")
	}

	// 尝试再次接受连接，应该失败
	_, err = listener.Accept()
	if err == nil {
		t.Fatal("Expected Accept to fail after listener closed")
	}
}

func TestTCPTransport_ConnectionInterface(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "TCP",
		Transport:         NewTCPTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "tcp-echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer SafeClose(clientConn, "tcp-client-connection")

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

	t.Log("TCP connection interface test passed")
}
