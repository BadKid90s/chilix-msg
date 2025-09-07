package transport

import (
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestWebSocketTransport_Protocol(t *testing.T) {
	transport := NewWebSocketTransport()
	if transport.Protocol() != "websocket" {
		t.Errorf("Expected protocol to be 'websocket', got '%s'", transport.Protocol())
	}
}

func TestWebSocketTransport_BasicConnection(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "WebSocket",
		Transport:         NewWebSocketTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunBasicConnectionTest(config)
	if !result.Success {
		t.Fatalf("Basic connection test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("WebSocket basic connection test: %s", result.Message)
}

func TestWebSocketTransport_LargeData(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "WebSocket",
		Transport:         NewWebSocketTransport(),
		ConnectionTimeout: 10 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunLargeDataTest(config)
	if !result.Success {
		t.Fatalf("Large data test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("WebSocket large data test: %s", result.Message)
}

func TestWebSocketTransport_Timeout(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "WebSocket",
		Transport:         NewWebSocketTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunTimeoutTest(config)
	if !result.Success {
		t.Fatalf("Timeout test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("WebSocket timeout test: %s", result.Message)
}

func TestWebSocketTransport_InvalidAddress(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "WebSocket",
		Transport:         NewWebSocketTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunInvalidAddressTest(t, config)
	if !result.Success {
		t.Fatalf("Invalid address test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("WebSocket invalid address test: %s", result.Message)
}

func TestWebSocketListener_Addr(t *testing.T) {
	transport := NewWebSocketTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer SafeClose(listener, "websocket-listener")

	addr := listener.Addr()
	if addr == nil {
		t.Fatal("Expected non-nil address")
	}

	// WebSocket监听器底层使用TCP，所以网络类型应该是tcp
	if addr.Network() != "tcp" {
		t.Errorf("Expected network to be 'tcp', got '%s'", addr.Network())
	}
}

func TestWebSocketListener_Close(t *testing.T) {
	transport := NewWebSocketTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// 在协程中关闭监听器，单独处理错误
	SafeClose(listener, "websocket-listener")

	// 等待一下让关闭操作完成
	time.Sleep(100 * time.Millisecond)

	// 尝试接受连接，应该失败
	_, err = listener.Accept()
	if err == nil {
		t.Fatal("Expected Accept to fail after listener closed")
	}
}

func TestWebSocketURLHandling(t *testing.T) {
	transport := NewWebSocketTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer SafeClose(listener, "websocket-listener")

	// 获取实际分配的地址
	addr := listener.Addr().String()

	// 验证地址格式
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Failed to split host:port: %v", err)
	}

	// 检查URL格式是否正确
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	if !strings.HasPrefix(u.String(), "ws://"+host) {
		t.Errorf("Expected URL to start with 'ws://%s', got '%s'", host, u.String())
	}
}

func TestWebSocketTransport_ConnectionInterface(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "WebSocket",
		Transport:         NewWebSocketTransport(),
		ConnectionTimeout: 5 * time.Second,
		DataTimeout:       2 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "websocket-echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer SafeClose(clientConn, "websocket-client-connection")

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

	t.Log("WebSocket connection interface test passed")
}
