package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/BadKid90s/chilix-msg/core"
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
	defer listener.Close()

	fmt.Println("✅ 服务器启动在 :9999")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept connection failed: %v", err)
			continue
		}

		go handleServerConnection(conn)
	}
}

func handleServerConnection(conn net.Conn) {
	defer conn.Close()

	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})
	defer processor.Close()

	// 注册正常处理器
	processor.RegisterHandler("get_data", func(ctx core.Context) error {
		var request map[string]interface{}
		if err := ctx.Bind(&request); err != nil {
			return ctx.Error("无效的请求格式")
		}

		userID, ok := request["user_id"]
		if !ok {
			// 这里调用 ctx.Error() 现在会正确发送错误响应
			return ctx.Error("缺少 user_id 参数")
		}

		// 模拟用户不存在的情况
		if userID == "999" {
			return ctx.Error("用户不存在")
		}

		// 正常返回数据
		return ctx.Reply(map[string]interface{}{
			"user_id": userID,
			"name":    "张三",
			"age":     25,
		})
	})

	fmt.Println("📡 服务器处理连接")
	if err := processor.Listen(); err != nil {
		log.Printf("Connection error: %v", err)
	}
}

func startClient() {
	tr := transport.NewTCPTransport()
	conn, err := tr.Dial("127.0.0.1:9999")
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer conn.Close()

	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   5 * time.Second,
	})
	defer processor.Close()

	// 启动客户端监听
	go func() {
		processor.Listen()
	}()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("🚀 客户端连接成功")

	// 测试 1: 正常请求
	fmt.Println("\n📤 测试正常请求:")
	resp, err := processor.Request("get_data", map[string]interface{}{
		"user_id": "123",
	})
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
	} else if resp.IsError() {
		fmt.Printf("❌ 服务器错误: %s\n", resp.Error())
	} else {
		var user map[string]interface{}
		resp.Bind(&user)
		fmt.Printf("✅ 获取用户成功: %+v\n", user)
	}

	// 测试 2: 缺少参数的请求
	fmt.Println("\n📤 测试缺少参数的请求:")
	resp, err = processor.Request("get_data", map[string]interface{}{})
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
	} else if resp.IsError() {
		fmt.Printf("✅ 正确接收到错误响应: %s\n", resp.Error())
	} else {
		fmt.Printf("❓ 意外的成功响应\n")
	}

	// 测试 3: 用户不存在的请求
	fmt.Println("\n📤 测试用户不存在的请求:")
	resp, err = processor.Request("get_data", map[string]interface{}{
		"user_id": "999",
	})
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
	} else if resp.IsError() {
		fmt.Printf("✅ 正确接收到错误响应: %s\n", resp.Error())
	} else {
		fmt.Printf("❓ 意外的成功响应\n")
	}

	fmt.Println("\n🎉 错误处理演示完成！")
	fmt.Println("现在 Writer.Error() 方法能够正确发送错误响应而不是错误消息。")
}