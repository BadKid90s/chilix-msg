# chilix-msg

chilix-msg 是一个高性能、轻量级的消息处理框架，专为分布式系统和微服务架构设计。它提供了简洁的API来处理消息通信，支持中间件、请求-响应模式，并且可以轻松扩展。

## 特性

- 🚀 高性能消息处理
- 🧩 模块化设计
- 🔌 支持多种传输协议（TCP、WebSocket等）
- 🛡️ 内置中间件支持（日志、恢复、加密等）
- 📦 支持多种序列化格式（JSON、Msgpack等）
- 🔄 请求-响应模式
- 🔒 端到端加密支持

## 安装
```bash
go get github.com/BadKid90s/chilix-msg
```

## 快速开始

### 服务器端
```go
package main

import (
	"log"
	"net"
	"time"

	"github.com/BadKid90s/chilix-msg/pkg/core"
	"github.com/BadKid90s/chilix-msg/pkg/serializer"
	"github.com/BadKid90s/chilix-msg/pkg/transport"
)

func main() {

	// 创建TCP传输
	

	listener, err := tcpTransport.Listen(":8080")

	if err != nil {

		log.Fatal("Server start failed:", err)

	}

	defer listener.Close()

	log.Println("✅ Server started on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept connection failed:", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	// 创建处理器

	processor := core.NewProcessor(conn, core.ProcessorOptions{

		Serializer: serializer.NewJSON(),

		MessageSizeLimit: 1024 * 1024, // 1MB

		RequestTimeout: 10 * time.Second,
	})

	// 注册中间件
	processor.Use(handler.LoggingMiddleware())
	processor.Use(handler.RecoveryMiddleware())

	// 注册消息处理器
	processor.RegisterHandler("login", func(ctx core.Context) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := ctx.Bind(&req); err != nil {
			return ctx.Error("Invalid request format")
		}

		log.Printf("Login request from %s", req.Username)

		// 处理登录逻辑
		if req.Username == "admin" && req.Password == "123456" {
			return ctx.Writer().Reply(ctx.RequestID(), "login_response", map[string]interface{}{
				"success": true,
				"user_id": 1001,
			})
		}

		return ctx.Writer().Reply(ctx.RequestID(), "login_response", map[string]interface{}{
			"success": false,
			"error":   "Invalid credentials",
		})
	})

	// 启动监听
	if err := processor.Listen(); err != nil {
		log.Printf("Connection error: %v", err)
	}

}

```

### 客户端
