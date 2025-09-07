package transport

import (
	"testing"
	"time"
)

func TestKCPTransport_Protocol(t *testing.T) {
	transport := NewKCPTransport()
	if transport.Protocol() != "kcp" {
		t.Errorf("Expected protocol to be 'kcp', got '%s'", transport.Protocol())
	}
}

func TestKCPTransport_BasicConnection(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "KCP",
		Transport:         NewKCPTransport(),
		ConnectionTimeout: 15 * time.Second, // KCP需要更长时间建立连接
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunBasicConnectionTest(config)
	if !result.Success {
		t.Fatalf("Basic connection test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("KCP basic connection test: %s", result.Message)
}

func TestKCPTransport_LargeData(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "KCP",
		Transport:         NewKCPTransport(),
		ConnectionTimeout: 20 * time.Second, // KCP需要更长时间处理大数据
		DataTimeout:       10 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunLargeDataTest(config)
	if !result.Success {
		t.Fatalf("Large data test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("KCP large data test: %s", result.Message)
}

func TestKCPTransport_Timeout(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "KCP",
		Transport:         NewKCPTransport(),
		ConnectionTimeout: 10 * time.Second,
		DataTimeout:       3 * time.Second,
		SkipNetworkTests:  false,
	}

	result := RunTimeoutTest(config)
	if !result.Success {
		t.Fatalf("Timeout test failed: %s - %v", result.Message, result.Error)
	}
	t.Logf("KCP timeout test: %s", result.Message)
}

func TestKCPTransport_InvalidAddress(t *testing.T) {
	transport := NewKCPTransport()

	// 测试无效地址格式
	_, err := transport.Dial("invalid-address")
	if err == nil {
		t.Fatal("Expected dial to fail with invalid address")
	}

	// 对于KCP，我们只测试无效地址格式，不测试不存在的服务器
	// 因为KCP可能需要更长时间来检测连接失败，这在测试环境中可能不稳定
	t.Log("KCP invalid address test passed")
}

func TestKCPListener_Addr(t *testing.T) {
	transport := NewKCPTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer SafeClose(listener, "kcp-listener")

	addr := listener.Addr()
	if addr == nil {
		t.Fatal("Expected non-nil address")
	}

	// KCP监听器底层使用UDP，所以网络类型应该是udp
	if addr.Network() != "udp" {
		t.Errorf("Expected network to be 'udp', got '%s'", addr.Network())
	}
}

func TestKCPListener_Close(t *testing.T) {
	transport := NewKCPTransport()

	// 使用随机端口启动监听器
	listener, err := transport.Listen("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	// 在协程中关闭监听器，单独处理错误
	SafeClose(listener, "kcp-listener")

	// 等待一下让关闭操作完成
	time.Sleep(100 * time.Millisecond)

	// 尝试接受连接，应该失败
	_, err = listener.Accept()
	if err == nil {
		t.Fatal("Expected Accept to fail after listener closed")
	}
}

func TestKCPTransport_ConnectionInterface(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "KCP",
		Transport:         NewKCPTransport(),
		ConnectionTimeout: 15 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "kcp-echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer SafeClose(clientConn, "kcp-client-connection")

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

	t.Log("KCP connection interface test passed")
}

func TestKCPTransport_Reliability(t *testing.T) {
	config := TestConfig{
		ProtocolName:      "KCP",
		Transport:         NewKCPTransport(),
		ConnectionTimeout: 15 * time.Second,
		DataTimeout:       5 * time.Second,
		SkipNetworkTests:  false,
	}

	// 创建回显服务器
	server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer SafeClose(server, "kcp-echo-server")

	// 客户端连接
	clientConn, err := config.Transport.Dial(server.Addr().String())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer SafeClose(clientConn, "kcp-client-connection")

	// 设置连接超时
	if err := clientConn.SetDeadline(time.Now().Add(config.ConnectionTimeout)); err != nil {
		t.Error(err.Error())
	}

	// 测试多次数据传输，验证KCP的可靠性
	for i := 0; i < 10; i++ {
		testData := []byte("KCP reliability test message " + string(rune('0'+i)))

		// 发送数据
		_, err = clientConn.Write(testData)
		if err != nil {
			t.Fatalf("Failed to write data (iteration %d): %v", i, err)
		}

		// 接收回显数据
		buf := make([]byte, len(testData))
		_, err = clientConn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read data (iteration %d): %v", i, err)
		}

		// 验证数据
		if string(buf) != string(testData) {
			t.Errorf("Data mismatch (iteration %d): expected %s, got %s", i, testData, buf)
		}
	}

	t.Log("KCP reliability test passed")
}
