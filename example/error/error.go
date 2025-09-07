package main

import (
	"net"
	"time"

	"github.com/BadKid90s/chilix-msg/core"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
	"github.com/BadKid90s/chilix-msg/transport"
)

func main() {
	// 启动服务器
	go startServer()
	time.Sleep(100 * time.Millisecond)

	// 启动客户端
	startClient()
}

func startServer() {
	tr := transport.NewTCPTransport()
	listener, err := tr.Listen("127.0.0.1:9999")
	if err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Errorf("Error closing listener: %v", err)
		}
	}()

	log.Infof("✅ 服务器启动在 :9999")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Accept connection failed: %v", err)
			continue
		}

		go handleServerConnection(conn)
	}
}

func handleServerConnection(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Errorf("Error closing connection: %v", err)
		}
	}()

	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})
	defer func() {
		err := processor.Close()
		if err != nil {
			log.Errorf("Error closing processor: %v", err)
		}
	}()

	// 注册正常处理器
	processor.RegisterHandler("get_data", func(ctx core.Context) error {

		var request map[string]interface{}
		if err := ctx.Bind(&request); err != nil {
			// 通信协议错误，返回错误响应
			return ctx.Reply(map[string]string{"error": "无效的请求格式"})
		}

		userID, ok := request["user_id"]
		if !ok {
			// 业务逻辑错误，返回自定义错误结构
			return ctx.Reply(map[string]interface{}{
				"success": false,
				"error":   "缺少 user_id 参数",
				"code":    "MISSING_PARAMETER",
			})
		}

		// 模拟用户不存在的情况
		if userID == "999" {
			// 业务逻辑错误，返回自定义错误结构
			return ctx.Reply(map[string]interface{}{
				"success": false,
				"error":   "用户不存在",
				"code":    "USER_NOT_FOUND",
			})
		}

		// 正常返回数据
		return ctx.Reply(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"user_id": userID,
				"name":    "张三",
				"age":     25,
			},
		})
	})

	log.Infof("📡 服务器处理连接")
	if err := processor.Listen(); err != nil {
		log.Errorf("Connection error: %v", err)
	}
}

func startClient() {
	tr := transport.NewTCPTransport()
	conn, err := tr.Dial("127.0.0.1:9999")
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			log.Errorf("Error closing connection: %v", err)
		}
	}(conn)

	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
	})
	defer func(processor core.Processor) {
		err := processor.Close()
		if err != nil {
			log.Errorf("Error closing processor: %v", err)
		}
	}(processor)

	// 启动客户端监听
	go func() {
		err2 := processor.Listen()
		if err != nil {
			log.Errorf("Error listening: %v", err2)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	log.Infof("🚀 客户端连接成功")

	// 测试 1: 正常请求
	log.Infof("\n📤 测试正常请求:")
	resp, err := processor.Request("get_data", map[string]interface{}{
		"user_id": "123",
	})
	if err != nil {
		log.Infof("❌ 通信错误: %v\n", err)
	} else {
		var result map[string]interface{}
		if err := resp.Bind(&result); err != nil {
			log.Infof("Error binding response: %v\n", err)
		}
		if success, ok := result["success"].(bool); ok && success {
			log.Infof("✅ 获取用户成功: %+v\n", result["data"])
		} else {
			log.Infof("❌ 业务错误: %s (错误码: %s)\n", result["error"], result["code"])
		}
	}

	// 测试 2: 缺少参数的请求
	log.Infof("\n📤 测试缺少参数的请求:")
	resp, err = processor.Request("get_data", map[string]interface{}{})
	if err != nil {
		log.Infof("❌ 通信错误: %v\n", err)
	} else {
		var result map[string]interface{}
		err := resp.Bind(&result)
		if err != nil {
			log.Infof("Error binding response: %v\n", err)
		}
		if success, ok := result["success"].(bool); ok && success {
			log.Infof("❓ 意外的成功响应\n")
		} else {
			log.Infof("✅ 正确处理业务错误: %s (错误码: %s)\n", result["error"], result["code"])
		}
	}

	// 测试 3: 用户不存在的请求
	log.Infof("\n📤 测试用户不存在的请求:")
	resp, err = processor.Request("get_data", map[string]interface{}{
		"user_id": "999",
	})
	if err != nil {
		log.Infof("❌ 通信错误: %v\n", err)
	} else {
		var result map[string]interface{}
		if err := resp.Bind(&result); err != nil {
			log.Errorf("Error binding response: %v\n", err)
		}
		if success, ok := result["success"].(bool); ok && success {
			log.Infof("❓ 意外的成功响应\n")
		} else {
			log.Infof("✅ 正确处理业务错误: %s (错误码: %s)\n", result["error"], result["code"])
		}
	}

	// 测试 4: 通信错误(发送到不存在的消息类型)
	log.Infof("\n📤 测试通信错误(超时):")
	_, err = processor.Request("non_existent_handler", map[string]interface{}{
		"data": "test",
	})
	if err != nil {
		log.Infof("✅ 正确捕获通信错误: %v\n", err)
	} else {
		log.Infof("❓ 意外收到响应\n")
	}

	log.Infof("\n🎉 错误处理演示完成！")
	log.Infof("📝 正确的错误处理方式:")
	log.Infof("   - 通信错误: 通过 err 返回值处理")
	log.Infof("   - 业务错误: 通过自定义响应结构处理")
	log.Infof("   - 框架不干涉业务逻辑，保持纯粹性")
}
