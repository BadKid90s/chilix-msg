package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/BadKid90s/chilix-msg/pkg/core"
	"github.com/BadKid90s/chilix-msg/pkg/serializer"
	"github.com/BadKid90s/chilix-msg/pkg/transport"
)

const (
	Port = ":8080"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2) // 等待服务器和客户端完成

	// 启动服务器
	go func() {
		defer wg.Done()
		startServer()
	}()

	// 给服务器一点时间启动
	time.Sleep(100 * time.Millisecond)

	// 启动客户端
	go func() {
		defer wg.Done()
		startClient()
	}()

	// 等待所有goroutine完成
	wg.Wait()
	fmt.Println("✅ All tasks completed")
}

func startServer() {
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Echo server start failed: %v", err)
	}
	defer listener.Close()

	log.Printf("✅ Echo server started on %s", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept connection failed: %v", err)
			continue
		}

		go handleEchoServerConnection(conn)
	}
}

func startClient() {
	conn, err := net.Dial("tcp", "localhost"+Port)
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer conn.Close()

	handleEchoClientConnection(conn)
}

func handleEchoServerConnection(conn transport.Connection) {
	defer conn.Close()
	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})

	// 注册消息处理器
	processor.RegisterHandler("echo", func(ctx core.Context) error {
		var msg string
		if err := ctx.Bind(&msg); err != nil {
			return ctx.Error("Invalid message format")
		}

		log.Printf("Received echo request: %s", msg)

		// 使用相同的消息类型回复
		return ctx.Reply(msg)
	})

	// 启动监听
	if err := processor.Listen(); err != nil {
		log.Printf("Connection error: %v", err)
	}
}

func handleEchoClientConnection(conn transport.Connection) {
	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})

	// 启动监听
	go func() {
		if err := processor.Listen(); err != nil {
			log.Printf("Client listen error: %v", err)
		}
	}()

	// 发送10个echo请求
	for i := 1; i <= 10; i++ {
		msg := fmt.Sprintf("Echo-%d", i)
		response, err := processor.Request("echo", msg)
		if err != nil {
			log.Printf("Echo request failed: %v", err)
			continue
		}

		// 检查错误响应
		if response.IsError() {
			log.Printf("Error response: %s", response.Error())
			continue
		}

		var echoResponse string
		err = response.Bind(&echoResponse)
		if err != nil {
			log.Printf("Failed to parse echo response: %v", err)
			continue
		}
		log.Printf("Echo response: %s", echoResponse)
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("✅ Sent and received 10 echo messages")
}
