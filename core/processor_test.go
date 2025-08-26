package core

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessorWithAllTransports 测试所有传输协议下的 Processor 功能
func TestProcessorWithAllTransports(t *testing.T) {
	testCases := []struct {
		name      string
		transport transport.Transport
		skip      bool
		reason    string
	}{
		{
			name:      "TCP",
			transport: transport.NewTCPTransport(),
			skip:      false,
		},
		{
			name:      "WebSocket",
			transport: transport.NewWebSocketTransport(),
			skip:      false,
		},
		{
			name:      "KCP",
			transport: transport.NewKCPTransport(),
			skip:      false,
		},
		{
			name:      "QUIC", 
			transport: transport.NewQUICTransport(),
			skip:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skipf("Skipping %s: %s", tc.name, tc.reason)
			}

			// 测试基本的消息发送和接收
			t.Run("BasicMessageSendReceive", func(t *testing.T) {
				testBasicMessageSendReceive(t, tc.transport)
			})

			// 测试请求-响应模式
			t.Run("RequestResponse", func(t *testing.T) {
				testRequestResponse(t, tc.transport)
			})

			// 测试中间件功能
			t.Run("Middleware", func(t *testing.T) {
				testMiddleware(t, tc.transport)
			})

			// 测试并发处理
			t.Run("ConcurrentProcessing", func(t *testing.T) {
				testConcurrentProcessing(t, tc.transport)
			})

			// 测试错误处理
			t.Run("ErrorHandling", func(t *testing.T) {
				testErrorHandling(t, tc.transport)
			})

			// 测试超时处理
			t.Run("RequestTimeout", func(t *testing.T) {
				testRequestTimeout(t, tc.transport)
			})
		})
	}
}

// testBasicMessageSendReceive 测试基本的消息发送和接收
func testBasicMessageSendReceive(t *testing.T, tr transport.Transport) {
	// 创建服务端监听器
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	messageReceived := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	// 启动服务端
	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		processor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer processor.Close()

		processor.RegisterHandler("test_message", func(ctx Context) error {
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			messageReceived <- msg
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- processor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	// 等待服务端准备就绪
	time.Sleep(200 * time.Millisecond)

	// 创建客户端连接
	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 发送消息
	testMsg := "Hello from client!"
	err = clientProcessor.Send("test_message", testMsg)
	require.NoError(t, err)

	// 验证消息接收
	select {
	case receivedMsg := <-messageReceived:
		assert.Equal(t, testMsg, receivedMsg)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	wg.Wait()
}

// testRequestResponse 测试请求-响应模式
func testRequestResponse(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动服务端
	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		serverProcessor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer serverProcessor.Close()

		// 注册echo处理器
		serverProcessor.RegisterHandler("echo", func(ctx Context) error {
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			// 回复相同的消息
			return ctx.Reply("ECHO: " + msg)
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- serverProcessor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(12 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 启动客户端监听
	go func() {
		clientProcessor.Listen()
	}()
	time.Sleep(100 * time.Millisecond)

	// 发送请求并等待响应
	testMsg := "Hello Server!"
	resp, err := clientProcessor.Request("echo", testMsg)
	require.NoError(t, err)

	var responseMsg string
	err = resp.Bind(&responseMsg)
	require.NoError(t, err)

	assert.Equal(t, "ECHO: "+testMsg, responseMsg)

	wg.Wait()
}

// testMiddleware 测试中间件功能
func testMiddleware(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	middlewareExecuted := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		serverProcessor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer serverProcessor.Close()

		// 注册测试中间件
		serverProcessor.Use(func(next Handler) Handler {
			return func(ctx Context) error {
				middlewareExecuted <- true
				return next(ctx)
			}
		})

		serverProcessor.RegisterHandler("test", func(ctx Context) error {
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- serverProcessor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 发送消息触发中间件
	err = clientProcessor.Send("test", "middleware test")
	require.NoError(t, err)

	// 验证中间件被执行
	select {
	case executed := <-middlewareExecuted:
		assert.True(t, executed)
	case <-time.After(5 * time.Second):
		t.Fatal("Middleware was not executed")
	}

	wg.Wait()
}

// testConcurrentProcessing 测试并发处理
func testConcurrentProcessing(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	messageCount := make(chan int, 100)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		serverProcessor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer serverProcessor.Close()

		serverProcessor.RegisterHandler("concurrent_test", func(ctx Context) error {
			var id int
			if err := ctx.Bind(&id); err != nil {
				return err
			}
			messageCount <- id
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- serverProcessor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(15 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 启动客户端监听
	go func() {
		clientProcessor.Listen()
	}()
	time.Sleep(100 * time.Millisecond)

	// 并发发送多个消息
	const numMessages = 10
	var clientWg sync.WaitGroup
	clientWg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func(id int) {
			defer clientWg.Done()
			// 添加小延迟避免并发写入竞争
			time.Sleep(time.Duration(id*10) * time.Millisecond)
			err := clientProcessor.Send("concurrent_test", id)
			if err != nil {
				t.Errorf("Send failed: %v", err)
			}
		}(i)
	}

	clientWg.Wait()

	// 验证所有消息都被接收
	received := make(map[int]bool)
	for i := 0; i < numMessages; i++ {
		select {
		case id := <-messageCount:
			received[id] = true
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent messages")
		}
	}

	assert.Equal(t, numMessages, len(received))

	wg.Wait()
}

// testErrorHandling 测试错误处理
func testErrorHandling(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		serverProcessor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer serverProcessor.Close()

		// 注册一个会返回错误的处理器
		serverProcessor.RegisterHandler("error_test", func(ctx Context) error {
			return ctx.Reply(map[string]string{"error": "test error message"})
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- serverProcessor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	go func() {
		clientProcessor.Listen()
	}()
	time.Sleep(100 * time.Millisecond)

	// 发送请求触发错误
	resp, err := clientProcessor.Request("error_test", "trigger error")
	require.NoError(t, err)

	// 验证错误响应
	require.NoError(t, err)
	// 检查响应类型是否为 "error_test"(因为我们使用的是 ctx.Reply)
	assert.Equal(t, "error_test", resp.MsgType())
	// 检查错误消息内容
	var errorData map[string]string
	err = resp.Bind(&errorData)
	require.NoError(t, err)
	assert.Equal(t, "test error message", errorData["error"])

	wg.Wait()
}

// testRequestTimeout 测试请求超时
func testRequestTimeout(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		serverProcessor := NewProcessor(conn, ProcessorOptions{
			Serializer:       serializer.DefaultSerializer,
			MessageSizeLimit: 1024 * 1024,
			RequestTimeout:   5 * time.Second,
			Logger:           log.NewDefaultLogger(),
		})
		defer serverProcessor.Close()

		// 注册一个永不响应的处理器
		serverProcessor.RegisterHandler("timeout_test", func(ctx Context) error {
			// 不发送响应，模拟超时
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- serverProcessor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   1 * time.Second, // 短超时时间
		Logger:           log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	go func() {
		clientProcessor.Listen()
	}()
	time.Sleep(100 * time.Millisecond)

	// 发送请求并期望超时
	_, err = clientProcessor.Request("timeout_test", "this will timeout")
	assert.Error(t, err)
	assert.Equal(t, ErrRequestTimeout, err)

	wg.Wait()
}

// TestProcessorAdvancedFeatures 测试处理器高级功能
func TestProcessorAdvancedFeatures(t *testing.T) {
	transport := transport.NewTCPTransport()
	
	// 测试消息大小限制
	t.Run("MessageSizeLimit", func(t *testing.T) {
		testMessageSizeLimit(t, transport)
	})

	// 测试序列化器配置
	t.Run("CustomSerializer", func(t *testing.T) {
		testCustomSerializer(t, transport)
	})

	// 测试处理器生命周期
	t.Run("ProcessorLifecycle", func(t *testing.T) {
		testProcessorLifecycle(t, transport)
	})
}

// testMessageSizeLimit 测试消息大小限制
func testMessageSizeLimit(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	messageReceived := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		processor := NewProcessor(conn, ProcessorOptions{
			MessageSizeLimit: 10, // 很小的限制
			Logger:           log.NewDefaultLogger(),
		})
		defer processor.Close()

		processor.RegisterHandler("test", func(ctx Context) error {
			messageReceived <- true
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- processor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Logger: log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 发送超大消息（会被服务端忽略）
	longMessage := strings.Repeat("a", 1000)
	err = clientProcessor.Send("test", longMessage)
	require.NoError(t, err)

	// 验证消息未被处理（由于大小限制）
	select {
	case <-messageReceived:
		t.Fatal("Message should not have been processed due to size limit")
	case <-time.After(2 * time.Second):
		// 预期的行为：消息被忽略
	}

	wg.Wait()
}

// testCustomSerializer 测试自定义序列化器
func testCustomSerializer(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	messageReceived := make(chan map[string]interface{}, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	customSerializer := &serializer.JSON{}

	go func() {
		defer wg.Done()
		
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("Accept failed: %v", err)
			return
		}
		defer conn.Close()

		processor := NewProcessor(conn, ProcessorOptions{
			Serializer: customSerializer,
			Logger:     log.NewDefaultLogger(),
		})
		defer processor.Close()

		processor.RegisterHandler("json_test", func(ctx Context) error {
			var data map[string]interface{}
			if err := ctx.Bind(&data); err != nil {
				return err
			}
			messageReceived <- data
			return nil
		})

		// 使用超时避免无限阻塞
		done := make(chan error, 1)
		go func() {
			done <- processor.Listen()
		}()
		
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Log("Server listen timeout")
		}
	}()

	time.Sleep(200 * time.Millisecond)

	clientConn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	defer clientConn.Close()

	clientProcessor := NewProcessor(clientConn, ProcessorOptions{
		Serializer: customSerializer,
		Logger:     log.NewDefaultLogger(),
	})
	defer clientProcessor.Close()

	// 发送JSON数据
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
	}
	err = clientProcessor.Send("json_test", testData)
	require.NoError(t, err)

	// 验证数据接收
	select {
	case receivedData := <-messageReceived:
		assert.Equal(t, testData["name"], receivedData["name"])
		assert.Equal(t, float64(25), receivedData["age"]) // JSON数字解析为float64
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	wg.Wait()
}

// testProcessorLifecycle 测试处理器生命周期
func testProcessorLifecycle(t *testing.T, tr transport.Transport) {
	listener, err := tr.Listen("127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			defer conn.Close()
			processor := NewProcessor(conn, ProcessorOptions{
				Logger: log.NewDefaultLogger(),
			})

			// 测试基本方法
			assert.NotNil(t, processor.Logger())
			assert.NotNil(t, processor.Serializer())
			
			// 测试关闭
			err := processor.Close()
			assert.NoError(t, err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	conn, err := tr.Dial(listener.Addr().String())
	require.NoError(t, err)
	conn.Close()
}