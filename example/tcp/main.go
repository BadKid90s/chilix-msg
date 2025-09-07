package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/BadKid90s/chilix-msg/core"
	"github.com/BadKid90s/chilix-msg/log"
	"github.com/BadKid90s/chilix-msg/serializer"
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
	log.Infof("✅ All tasks completed")
}

func startServer() {
	listener, err := net.Listen("tcp", Port)
	if err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Infof("Error closing listener: %v", err)
		}
	}()

	log.Infof("✅ Server started on %s", listener.Addr())

	// 接受一个连接
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Accept connection failed: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Infof("Error closing connection: %v", err)
		}
	}()

	log.Infof("Client connected: %s", conn.RemoteAddr())

	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   3 * time.Second,
	})

	// 注册消息处理器
	processor.RegisterHandler("get_time", func(ctx core.Context) error {
		if err != nil {
			log.Infof("Error registering handler: %v", err)
		}
		currentTime := time.Now().Format(time.RFC3339)
		log.Infof("Received time request, sending response")
		return ctx.Reply(currentTime)
	})

	// 启动主动推送
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		counter := 0
		for range ticker.C {
			counter++
			update := fmt.Sprintf("Server update #%d at %s", counter, time.Now().Format(time.RFC3339))
			if err := processor.Send("server_update", update); err != nil {
				log.Infof("Failed to send update: %v", err)
				return
			}
			log.Infof("Sent server update: %s", update)
		}
	}()

	// 启动监听
	if err := processor.Listen(); err != nil {
		log.Infof("Connection error: %v", err)
	}
}

func startClient() {
	conn, err := net.Dial("tcp", "localhost"+Port)
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Errorf("Error closing listener: %v", err)
		}
	}(conn)

	log.Infof("✅ Connected to server")

	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorConfig{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})

	// 注册消息处理器
	processor.RegisterHandler("time_response", func(ctx core.Context) error {
		if err != nil {
			log.Infof("Error registering handler: %v", err)
		}
		var timeStr string
		if err := ctx.Bind(&timeStr); err != nil {
			log.Infof("Failed to parse time response: %v", err)
			return nil
		}
		log.Infof("⏰ Received time response: %s", timeStr)
		return nil
	})

	processor.RegisterHandler("server_update", func(ctx core.Context) error {
		var update string
		if err := ctx.Bind(&update); err != nil {
			log.Infof("Failed to parse server update: %v", err)
			return nil
		}
		log.Infof("📡 Received server update: %s", update)
		return nil
	})

	// 启动监听
	go func() {
		if err := processor.Listen(); err != nil {
			log.Infof("Client listen error: %v", err)
		}
	}()

	// 发送时间请求
	log.Infof("Sending time request...")
	response, err := processor.Request("get_time", nil)
	if err != nil {
		log.Infof("Time request failed: %v", err)
	} else {
		var timeStr string
		if err := response.Bind(&timeStr); err != nil {
			log.Infof("Failed to parse time response: %v", err)
		} else {
			log.Infof("⏰ Server time: %s", timeStr)
		}
	}

	// 等待服务器推送
	log.Infof("Waiting for server updates...")
	time.Sleep(30 * time.Second)
	log.Infof("✅ Finished receiving server updates")
}
