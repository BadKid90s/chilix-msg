package core

import (
	"fmt"
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

// TestHelper provides common testing utilities
type TestHelper struct {
	logger log.Logger
}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{
		logger: log.NewDefaultLogger(),
	}
}

// TestServer represents a test server
type TestServer struct {
	Listener  transport.Listener
	Processor Processor
	CloseOnce sync.Once
	Ready     chan struct{}
}

// TestClient represents a test client
type TestClient struct {
	Conn      transport.Connection
	Processor Processor
}

// StartServer starts a test server with the given transport and handler
func (h *TestHelper) StartServer(tr transport.Transport, handler func(Processor)) (*TestServer, error) {
	return h.StartServerWithConfig(tr, handler, ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   1 * time.Second,
		Logger:           h.logger,
	})
}

// StartServerWithConfig starts a test server with custom configuration
func (h *TestHelper) StartServerWithConfig(tr transport.Transport, handler func(Processor), opts ProcessorConfig) (*TestServer, error) {
	h.logger.Debugf("Starting test server with transport %T", tr)

	listener, err := tr.Listen("127.0.0.1:0")
	if err != nil {
		h.logger.Errorf("Failed to create listener: %v", err)
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	h.logger.Debugf("Server listening on %s", listener.Addr().String())

	server := &TestServer{
		Listener: listener,
		Ready:    make(chan struct{}),
	}

	// Channel to signal server encountered an error
	serverError := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		h.logger.Debugf("Waiting for incoming connection")
		conn, err := listener.Accept()
		if err != nil {
			h.logger.Errorf("Accept failed: %v", err)
			serverError <- fmt.Errorf("accept failed: %w", err)
			return
		}
		h.logger.Debugf("Accepted connection from %s", conn.RemoteAddr().String())

		processor := NewProcessor(conn, opts)

		server.Processor = processor

		// Run the handler
		if handler != nil {
			h.logger.Debugf("Running server handler")
			handler(processor)
		}

		// Notify that server is fully ready for use
		close(server.Ready)

		// Start listening
		h.logger.Debugf("Starting processor listen loop")
		if err := processor.Listen(); err != nil {
			h.logger.Errorf("Processor listen failed: %v", err)
		}
		h.logger.Debugf("Processor listen loop finished")
	}()

	// Non-blocking return - server will start asynchronously
	// Client should connect soon after this returns
	return server, nil
}

// CloseServer closes the test server
func (s *TestServer) Close() error {
	var err error
	s.CloseOnce.Do(func() {
		// Close processor first
		if s.Processor != nil {
			h := s.Processor.Logger()
			if h != nil {
				h.Debugf("Closing processor")
				if procErr := s.Processor.Close(); procErr != nil {
					h.Errorf("Failed to close processor: %v", procErr)
					err = fmt.Errorf("failed to close processor: %w", procErr)
				} else {
					h.Debugf("Processor closed successfully")
				}
			} else {
				// Fallback to default logger
				defaultLogger := log.NewDefaultLogger()
				defaultLogger.Debugf("Closing processor")
				if procErr := s.Processor.Close(); procErr != nil {
					defaultLogger.Errorf("Failed to close processor: %v", procErr)
					err = fmt.Errorf("failed to close processor: %w", procErr)
				} else {
					defaultLogger.Debugf("Processor closed successfully")
				}
			}
		}

		// Close listener
		if s.Listener != nil {
			defaultLogger := log.NewDefaultLogger()
			defaultLogger.Debugf("Closing listener")
			if listenerErr := s.Listener.Close(); listenerErr != nil {
				defaultLogger.Errorf("Failed to close listener: %v", listenerErr)
				if err == nil {
					err = fmt.Errorf("failed to close listener: %w", listenerErr)
				} else {
					err = fmt.Errorf("%v; failed to close listener: %w", err, listenerErr)
				}
			} else {
				defaultLogger.Debugf("Listener closed successfully")
			}
		}
	})
	return err
}

// CloseClient closes the test client
func (c *TestClient) Close() error {
	if c.Processor != nil {
		h := c.Processor.Logger()
		if h != nil {
			h.Debugf("Closing client processor")
			return c.Processor.Close()
		} else {
			// Fallback to default logger
			defaultLogger := log.NewDefaultLogger()
			defaultLogger.Debugf("Closing client processor")
			return c.Processor.Close()
		}
	}
	return c.Conn.Close()
}

// StartClient starts a test client with the given transport and address
func (h *TestHelper) StartClient(tr transport.Transport, addr string) (*TestClient, error) {
	return h.StartClientWithConfig(tr, addr, ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   1 * time.Second,
		Logger:           h.logger,
	})
}

// StartClientWithConfig starts a test client with custom configuration
func (h *TestHelper) StartClientWithConfig(tr transport.Transport, addr string, opts ProcessorConfig) (*TestClient, error) {
	h.logger.Debugf("Starting test client to connect to %s with transport %T", addr, tr)

	// Try to connect with timeout
	var conn transport.Connection
	var connectErr error

	done := make(chan struct{})
	go func() {
		h.logger.Debugf("Attempting to dial server at %s", addr)
		conn, connectErr = tr.Dial(addr)
		if connectErr != nil {
			h.logger.Errorf("Dial failed: %v", connectErr)
		} else {
			h.logger.Debugf("Connected to server at %s", conn.RemoteAddr().String())
		}
		close(done)
	}()

	select {
	case <-done:
		if connectErr != nil {
			return nil, fmt.Errorf("failed to dial: %w", connectErr)
		}
	case <-time.After(30 * time.Second): // 增加超时时间到30秒
		h.logger.Errorf("Client connect timeout after 30 seconds")
		return nil, fmt.Errorf("client connect timeout")
	}

	processor := NewProcessor(conn, opts)

	client := &TestClient{
		Conn:      conn,
		Processor: processor,
	}

	// Start client listener in a goroutine
	h.logger.Debugf("Starting client processor listen loop")
	go func() {
		if err := processor.Listen(); err != nil {
			h.logger.Errorf("Client processor listen failed: %v", err)
		}
		h.logger.Debugf("Client processor listen loop finished")
	}()

	return client, nil
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Operation string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout waiting for %s", e.Operation)
}

// TestError represents a general test error
type TestError struct {
	Message string
	Cause   error
}

func (e *TestError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error
func (e *TestError) Unwrap() error {
	return e.Cause
}

// WithTimeout runs a function with a timeout
func WithTimeout(fn func() error, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return &TimeoutError{Operation: "operation"}
	}
}

// WaitForServerReady waits for the server to be ready
func (s *TestServer) WaitForServerReady(timeout time.Duration) error {
	select {
	case <-s.Ready:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("server not ready within timeout")
	}
}

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
	helper := NewTestHelper()

	// 创建服务端
	messageReceived := make(chan string, 1)
	server, err := helper.StartServer(tr, func(p Processor) {
		p.RegisterHandler("test_message", func(ctx Context) error {
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			messageReceived <- msg
			return nil
		})
	})
	require.NoError(t, err, "Failed to start server")
	defer func() {
		err := server.Close()
		require.NoError(t, err)
	}()

	t.Logf("Server started successfully, address: %s", server.Listener.Addr().String())

	// 创建客户端
	t.Log("Attempting to start client...")
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err, "Failed to start client")
	defer func() {
		err := client.Close()
		require.NoError(t, err)
	}()

	t.Log("Client started successfully")

	// 发送消息
	testMsg := "Hello from client!"
	t.Logf("Sending message: %s", testMsg)
	err = client.Processor.Send("test_message", testMsg)
	require.NoError(t, err, "Failed to send message")

	t.Log("Message sent successfully")

	// 验证消息接收
	select {
	case receivedMsg := <-messageReceived:
		assert.Equal(t, testMsg, receivedMsg)
		t.Log("Message received successfully")
	case <-time.After(5 * time.Second): // 增加超时时间
		t.Fatal("Timeout waiting for message")
	}
}

// testRequestResponse 测试请求-响应模式
func testRequestResponse(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	server, err := helper.StartServer(tr, func(p Processor) {
		// 注册echo处理器
		p.RegisterHandler("echo", func(ctx Context) error {
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			// 回复相同的消息
			return ctx.Reply("ECHO: " + msg)
		})
	})
	require.NoError(t, err)
	defer func() {
		err := server.Close()
		require.NoError(t, err)
	}()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer func() {
		err := client.Close()
		require.NoError(t, err)
	}()

	// 发送请求并等待响应
	testMsg := "Hello Server!"
	resp, err := client.Processor.Request("echo", testMsg)
	require.NoError(t, err)

	var responseMsg string
	err = resp.Bind(&responseMsg)
	require.NoError(t, err)

	assert.Equal(t, "ECHO: "+testMsg, responseMsg)
}

// testMiddleware 测试中间件功能
func testMiddleware(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	middlewareExecuted := make(chan bool, 1)
	server, err := helper.StartServer(tr, func(p Processor) {
		// 注册测试中间件
		p.Use(func(next Handler) Handler {
			return func(ctx Context) error {
				middlewareExecuted <- true
				return next(ctx)
			}
		})

		p.RegisterHandler("test", func(ctx Context) error {
			return nil
		})
	})
	require.NoError(t, err)
	defer func() {
		err := server.Close()
		require.NoError(t, err)
	}()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer func() {
		err := client.Close()
		require.NoError(t, err)
	}()

	// 发送消息触发中间件
	err = client.Processor.Send("test", "middleware test")
	require.NoError(t, err)

	// 验证中间件被执行
	select {
	case executed := <-middlewareExecuted:
		assert.True(t, executed)
	case <-time.After(2 * time.Second): // 减少超时时间
		t.Fatal("Middleware was not executed")
	}
}

// testConcurrentProcessing 测试并发处理
func testConcurrentProcessing(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	messageCount := make(chan int, 100)
	server, err := helper.StartServer(tr, func(p Processor) {
		p.RegisterHandler("concurrent_test", func(ctx Context) error {
			var id int
			if err := ctx.Bind(&id); err != nil {
				return err
			}
			messageCount <- id
			return nil
		})
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 并发发送多个消息
	const numMessages = 10
	var clientWg sync.WaitGroup
	clientWg.Add(numMessages)

	for i := 0; i < numMessages; i++ {
		go func(id int) {
			defer clientWg.Done()
			// 添加小延迟避免并发写入竞争
			time.Sleep(time.Duration(id*5) * time.Millisecond) // 减少延迟时间
			err := client.Processor.Send("concurrent_test", id)
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
		case <-time.After(2 * time.Second): // 减少超时时间
			t.Fatal("Timeout waiting for concurrent messages")
		}
	}

	assert.Equal(t, numMessages, len(received))
}

// testErrorHandling 测试错误处理
func testErrorHandling(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	server, err := helper.StartServer(tr, func(p Processor) {
		// 注册一个会返回错误的处理器
		p.RegisterHandler("error_test", func(ctx Context) error {
			return ctx.Reply(map[string]string{"error": "test error message"})
		})
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 发送请求触发错误
	resp, err := client.Processor.Request("error_test", "trigger error")
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
}

// testRequestTimeout 测试请求超时
func testRequestTimeout(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	server, err := helper.StartServer(tr, func(p Processor) {
		// 注册一个永不响应的处理器
		p.RegisterHandler("timeout_test", func(ctx Context) error {
			// 不发送响应，模拟超时
			return nil
		})
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端，使用短超时时间
	conn, err := tr.Dial(server.Listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()

	clientProcessor := NewProcessor(conn, ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   500 * time.Millisecond, // 短超时时间
		Logger:           log.NewDefaultLogger(),
	})

	// 使用WaitGroup确保goroutine完成
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := clientProcessor.Listen()
		// 不再检查错误，因为在测试结束时关闭连接会产生预期的错误
		if err != nil {
			// 只记录错误，不使用require.NoError因为这会在测试结束后运行
			fmt.Printf("Client processor listen error: %v\n", err)
		}
	}()
	time.Sleep(50 * time.Millisecond) // 减少等待时间

	// 发送请求并期望超时
	_, err = clientProcessor.Request("timeout_test", "this will timeout")
	assert.Error(t, err)
	assert.Equal(t, ErrRequestTimeout, err)

	// 关闭处理器并等待goroutine完成
	err = clientProcessor.Close()
	assert.NoError(t, err)
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

	// 测试空消息处理
	t.Run("EmptyMessage", func(t *testing.T) {
		testEmptyMessage(t, transport)
	})

	// 测试大消息处理
	t.Run("LargeMessage", func(t *testing.T) {
		testLargeMessage(t, transport)
	})

	// 测试未注册的消息类型
	t.Run("UnregisteredMessageType", func(t *testing.T) {
		testUnregisteredMessageType(t, transport)
	})
}

// testMessageSizeLimit 测试消息大小限制
func testMessageSizeLimit(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端，使用很小的限制
	messageReceived := make(chan bool, 1)
	server, err := helper.StartServerWithConfig(tr, func(p Processor) {
		p.RegisterHandler("test", func(ctx Context) error {
			messageReceived <- true
			return nil
		})
	}, ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 10, // 很小的限制
		RequestTimeout:   1 * time.Second,
		Logger:           helper.logger,
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 发送超大消息（会被服务端忽略）
	longMessage := strings.Repeat("a", 1000)
	err = client.Processor.Send("test", longMessage)
	require.NoError(t, err)

	// 验证消息未被处理（由于大小限制）
	select {
	case <-messageReceived:
		t.Fatal("Message should not have been processed due to size limit")
	case <-time.After(2 * time.Second): // 减少超时时间
		// 预期的行为：消息被忽略
	}
}

// testCustomSerializer 测试自定义序列化器
func testCustomSerializer(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	messageReceived := make(chan map[string]interface{}, 1)
	customSerializer := &serializer.JSON{}

	server, err := helper.StartServerWithConfig(tr, func(p Processor) {
		p.RegisterHandler("json_test", func(ctx Context) error {
			var data map[string]interface{}
			if err := ctx.Bind(&data); err != nil {
				return err
			}
			messageReceived <- data
			return nil
		})
	}, ProcessorConfig{
		Serializer:       customSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   1 * time.Second,
		Logger:           helper.logger,
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClientWithConfig(tr, server.Listener.Addr().String(), ProcessorConfig{
		Serializer:       customSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   1 * time.Second,
		Logger:           helper.logger,
	})
	require.NoError(t, err)
	defer client.Close()

	// 发送JSON数据
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
	}
	err = client.Processor.Send("json_test", testData)
	require.NoError(t, err)

	// 验证数据接收
	select {
	case receivedData := <-messageReceived:
		assert.Equal(t, testData["name"], receivedData["name"])
		assert.Equal(t, float64(25), receivedData["age"]) // JSON数字解析为float64
	case <-time.After(2 * time.Second): // 减少超时时间
		t.Fatal("Timeout waiting for message")
	}
}

// testProcessorLifecycle 测试处理器生命周期
func testProcessorLifecycle(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	server, err := helper.StartServer(tr, func(p Processor) {
		// 测试基本方法
		assert.NotNil(t, p.Logger())
		assert.NotNil(t, p.Serializer())
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 等待服务器处理器初始化
	time.Sleep(100 * time.Millisecond)

	// 测试基本方法
	if server.Processor != nil {
		assert.NotNil(t, server.Processor.Logger())
		assert.NotNil(t, server.Processor.Serializer())
	}
}

// testEmptyMessage 测试空消息处理
func testEmptyMessage(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	messageReceived := make(chan bool, 1)
	server, err := helper.StartServer(tr, func(p Processor) {
		p.RegisterHandler("empty_test", func(ctx Context) error {
			// 检查是否为空消息
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			// 空字符串消息
			if msg == "" {
				messageReceived <- true
			}
			return nil
		})
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 发送空消息
	err = client.Processor.Send("empty_test", "")
	require.NoError(t, err)

	// 验证空消息被正确处理
	select {
	case <-messageReceived:
		// 正常处理
	case <-time.After(2 * time.Second): // 减少超时时间
		t.Fatal("Timeout waiting for empty message")
	}
}

// testLargeMessage 测试大消息处理
func testLargeMessage(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端
	messageReceived := make(chan int, 1)
	server, err := helper.StartServer(tr, func(p Processor) {
		p.RegisterHandler("large_test", func(ctx Context) error {
			var msg string
			if err := ctx.Bind(&msg); err != nil {
				return err
			}
			// 返回消息长度
			messageReceived <- len(msg)
			return nil
		})
	})
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 发送大消息 (10KB)
	largeMessage := strings.Repeat("a", 10*1024)
	err = client.Processor.Send("large_test", largeMessage)
	require.NoError(t, err)

	// 验证大消息被正确处理
	select {
	case length := <-messageReceived:
		assert.Equal(t, len(largeMessage), length)
	case <-time.After(2 * time.Second): // 减少超时时间
		t.Fatal("Timeout waiting for large message")
	}
}

// testUnregisteredMessageType 测试未注册的消息类型
func testUnregisteredMessageType(t *testing.T, tr transport.Transport) {
	helper := NewTestHelper()

	// 创建服务端（不注册任何处理器）
	server, err := helper.StartServer(tr, nil)
	require.NoError(t, err)
	defer server.Close()

	// 创建客户端
	client, err := helper.StartClient(tr, server.Listener.Addr().String())
	require.NoError(t, err)
	defer client.Close()

	// 发送未注册的消息类型
	err = client.Processor.Send("unregistered_type", "test message")
	require.NoError(t, err)

	// 服务器应该不会崩溃，消息会被忽略
	// 这里我们只是验证发送不报错，实际行为取决于实现
	time.Sleep(50 * time.Millisecond) // 减少等待时间
}
