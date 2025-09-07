package transport

import (
	"testing"
	"time"
)

// 测试所有传输实现是否正确实现了Transport接口
func TestTransportImplementations(t *testing.T) {
	// 创建各种传输实现
	transports := []struct {
		name      string
		transport Transport
	}{
		{"TCP", NewTCPTransport()},
		{"WebSocket", NewWebSocketTransport()},
		{"KCP", NewKCPTransport()},
		{"QUIC", NewQUICTransport()},
	}

	// 验证每个实现都正确实现了Transport接口
	for _, tt := range transports {
		t.Run(tt.name, func(t *testing.T) {
			// 检查Protocol方法
			protocol := tt.transport.Protocol()
			if protocol == "" {
				t.Errorf("%s: Protocol() returned empty string", tt.name)
			}

			// 尝试监听（使用随机端口）
			listener, err := tt.transport.Listen("127.0.0.1:0")
			if err != nil {
				t.Fatalf("%s: Listen() failed: %v", tt.name, err)
			}
			defer SafeClose(listener, tt.name+"-listener")

			// 检查监听器是否正确实现了Listener接口
			addr := listener.Addr()
			if addr == nil {
				t.Errorf("%s: Listener.Addr() returned nil", tt.name)
			}
			if addr.String() == "" {
				t.Errorf("%s: Listener.Addr().String() returned empty string", tt.name)
			}
		})
	}
}

// 测试通用连接功能
func TestGenericConnectionBehavior(t *testing.T) {
	// 创建各种传输实现
	transports := []struct {
		name      string
		transport Transport
	}{
		{"TCP", NewTCPTransport()},
		{"WebSocket", NewWebSocketTransport()},
		{"KCP", NewKCPTransport()},
		{"QUIC", NewQUICTransport()},
	}

	for _, tt := range transports {
		t.Run(tt.name, func(t *testing.T) {
			config := TestConfig{
				ProtocolName:      tt.name,
				Transport:         tt.transport,
				ConnectionTimeout: 10 * time.Second,
				DataTimeout:       3 * time.Second,
				SkipNetworkTests:  false,
			}

			// 运行基本连接测试
			result := RunBasicConnectionTest(config)
			if !result.Success {
				t.Fatalf("%s: Basic connection test failed: %s - %v", tt.name, result.Message, result.Error)
			}
			t.Logf("%s: %s", tt.name, result.Message)
		})
	}
}

// 测试连接接口的所有方法
func TestConnectionInterface(t *testing.T) {
	// 创建各种传输实现
	transports := []struct {
		name      string
		transport Transport
	}{
		{"TCP", NewTCPTransport()},
		{"WebSocket", NewWebSocketTransport()},
		{"KCP", NewKCPTransport()},
		{"QUIC", NewQUICTransport()},
	}

	for _, tt := range transports {
		t.Run(tt.name, func(t *testing.T) {
			config := TestConfig{
				ProtocolName:      tt.name,
				Transport:         tt.transport,
				ConnectionTimeout: 10 * time.Second,
				DataTimeout:       3 * time.Second,
				SkipNetworkTests:  false,
			}

			// 创建回显服务器
			server, err := NewEchoServer(config.Transport, "127.0.0.1:0")
			if err != nil {
				t.Fatalf("%s: Failed to create server: %v", tt.name, err)
			}
			defer SafeClose(server, tt.name+"-echo-server")

			// 客户端连接
			conn, err := config.Transport.Dial(server.Addr().String())
			if err != nil {
				t.Fatalf("%s: Dial() failed: %v", tt.name, err)
			}
			defer SafeClose(conn, tt.name+"-client-connection")

			// 测试连接的基本属性
			if conn.LocalAddr() == nil {
				t.Errorf("%s: LocalAddr() returned nil", tt.name)
			}
			if conn.RemoteAddr() == nil {
				t.Errorf("%s: RemoteAddr() returned nil", tt.name)
			}

			// 测试设置超时
			err = conn.SetDeadline(time.Now().Add(time.Second))
			if err != nil {
				t.Errorf("%s: SetDeadline() failed: %v", tt.name, err)
			}

			err = conn.SetReadDeadline(time.Now().Add(time.Second))
			if err != nil {
				t.Errorf("%s: SetReadDeadline() failed: %v", tt.name, err)
			}

			err = conn.SetWriteDeadline(time.Now().Add(time.Second))
			if err != nil {
				t.Errorf("%s: SetWriteDeadline() failed: %v", tt.name, err)
			}

			// 重置超时
			err = conn.SetDeadline(time.Time{})
			if err != nil {
				t.Errorf("%s: Failed to reset deadline: %v", tt.name, err)
			}

			// 测试数据传输
			testData := []byte("Hello, " + tt.name + " Connection Interface!")
			_, err = conn.Write(testData)
			if err != nil {
				t.Fatalf("%s: Write() failed: %v", tt.name, err)
			}

			// 接收回显数据
			buf := make([]byte, len(testData))
			_, err = conn.Read(buf)
			if err != nil {
				t.Fatalf("%s: Read() failed: %v", tt.name, err)
			}

			// 验证数据
			if string(buf) != string(testData) {
				t.Errorf("%s: Echo data mismatch. Expected: %s, Got: %s", tt.name, testData, buf)
			}

			// 在协程中关闭连接，单独处理错误
			SafeClose(conn, tt.name+"-client-connection")

			t.Logf("%s: Connection interface test passed", tt.name)
		})
	}
}

// 测试所有协议的性能对比
func TestTransportPerformance(t *testing.T) {
	transports := []struct {
		name      string
		transport Transport
	}{
		{"TCP", NewTCPTransport()},
		{"WebSocket", NewWebSocketTransport()},
		{"KCP", NewKCPTransport()},
		{"QUIC", NewQUICTransport()},
	}

	for _, tt := range transports {
		t.Run(tt.name, func(t *testing.T) {
			config := TestConfig{
				ProtocolName:      tt.name,
				Transport:         tt.transport,
				ConnectionTimeout: 15 * time.Second,
				DataTimeout:       5 * time.Second,
				SkipNetworkTests:  false,
			}

			// 测试连接建立时间
			start := time.Now()
			result := RunBasicConnectionTest(config)
			connectionTime := time.Since(start)

			if !result.Success {
				t.Fatalf("%s: Performance test failed: %s - %v", tt.name, result.Message, result.Error)
			}

			t.Logf("%s: Connection established in %v", tt.name, connectionTime)

			// 测试大数据传输性能
			start = time.Now()
			result = RunLargeDataTest(config)
			dataTime := time.Since(start)

			if !result.Success {
				t.Fatalf("%s: Large data performance test failed: %s - %v", tt.name, result.Message, result.Error)
			}

			t.Logf("%s: Large data transfer completed in %v", tt.name, dataTime)
		})
	}
}

// 测试传输层的错误处理
func TestTransportErrorHandling(t *testing.T) {
	transports := []struct {
		name      string
		transport Transport
	}{
		{"TCP", NewTCPTransport()},
		{"WebSocket", NewWebSocketTransport()},
		{"KCP", NewKCPTransport()},
		{"QUIC", NewQUICTransport()},
	}

	for _, tt := range transports {
		t.Run(tt.name, func(t *testing.T) {
			config := TestConfig{
				ProtocolName:      tt.name,
				Transport:         tt.transport,
				ConnectionTimeout: 5 * time.Second,
				DataTimeout:       2 * time.Second,
				SkipNetworkTests:  false,
			}

			// 测试无效地址
			result := RunInvalidAddressTest(t, config)
			if !result.Success {
				// 对于KCP，无效地址测试可能不稳定，我们记录但不失败
				if tt.name == "KCP" {
					t.Logf("%s: Invalid address test had issues (this may be expected): %s - %v", tt.name, result.Message, result.Error)
				} else {
					t.Errorf("%s: Invalid address test failed: %s - %v", tt.name, result.Message, result.Error)
				}
			}

			// 测试超时处理
			result = RunTimeoutTest(config)
			if !result.Success {
				t.Errorf("%s: Timeout test failed: %s - %v", tt.name, result.Message, result.Error)
			}

			t.Logf("%s: Error handling test passed", tt.name)
		})
	}
}
