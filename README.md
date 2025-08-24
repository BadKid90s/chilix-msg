# chilix-msg

chilix-msg 是一个高性能、轻量级的消息处理框架，专为分布式系统和微服务架构设计。它提供了简洁的API来处理消息通信，支持中间件、请求-响应模式，并且可以轻松扩展。

## 特性

- 🚀 **高性能消息处理**
  - 基于 goroutine 的并发处理模型
  - 高效的二进制协议编解码
  - 零拷贝消息处理优化

- 🧩 **模块化设计**
  - 可插拔的序列化器（JSON、Binary 等）
  - 灵活的中间件机制
  - 易于扩展的传输层接口

- 🔌 **支持多种传输协议**
  - TCP（当前默认支持）
  - KCP（计划中）
  - WebSocket（计划中）
  - 支持自定义传输协议

- 🛡️ **内置中间件支持**
  - 日志记录中间件
  - 异常恢复中间件
  - 端到端加密中间件
  - 支持自定义中间件

- 📦 **多种序列化格式**
  - JSON 序列化（默认）
  - Binary 序列化
  - 易于扩展的序列化接口

- 🔄 **请求-响应模式**
  - 同步请求-响应通信
  - 可配置的请求超时时间
  - 自动请求ID匹配机制

- 🔁 **消息推送**
  - 服务器主动推送消息
  - 支持广播和单播消息
  - 实时消息分发

- 🔒 **端到端加密支持**
  - AES-GCM 加密算法
  - 自动 nonce 生成
  - 透明的加解密处理

- ⚙️ **可配置的消息处理**
  - 消息大小限制
  - 超时控制
  - 错误处理机制

- 📊 **上下文管理**
  - 完整的消息上下文信息
  - 原始数据访问
  - 连接状态管理

- 📈 **可观测性**
  - 详细的日志记录
  - 性能指标统计
  - 错误追踪和恢复

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
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
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
func handleConnection(conn net.Conn) { // 创建处理器 processor := core.NewProcessor(conn, core.ProcessorOptions{ Serializer: serializer.DefaultSerializer, MessageSizeLimit: 1024 * 1024, // 1MB RequestTimeout: 10 * time.Second, })
    defer conn.Close()
	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   3 * time.Second,
	})

	// 注册消息处理器
	processor.RegisterHandler("get_time", func(ctx core.Context) error {
		currentTime := time.Now().Format(time.RFC3339)
		return ctx.Reply(currentTime)
	})

	// 启动监听
	if err := processor.Listen(); err != nil {
		log.Printf("Connection error: %v", err)
	}
}


```
### 客户端

```go
package main

import (
	"log"
	"net"
	"time"

	"github.com/BadKid90s/chilix-msg/pkg/core"
	"github.com/BadKid90s/chilix-msg/pkg/serializer"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Connect failed: %v", err)
	}
	defer conn.Close()
	// 创建处理器
	processor := core.NewProcessor(conn, core.ProcessorOptions{
		Serializer:       serializer.DefaultSerializer,
		MessageSizeLimit: 1024 * 1024,
		RequestTimeout:   10 * time.Second,
	})

	// 发送时间请求
	response, err := processor.Request("get_time", nil)
	if err != nil {
		log.Printf("Time request failed: %v", err)
	} else {
		var timeStr string
		if err := response.Bind(&timeStr); err != nil {
			log.Printf("Failed to parse time response: %v", err)
		} else {
			log.Printf("⏰ Server time: %s", timeStr)
		}
	}
}

```

## 协议格式

chilix-msg 使用基于长度前缀的二进制协议格式，确保消息的可靠传输和解析。

### 消息结构

每条消息由以下部分组成：
```text
+------------------+------------------+------------------+------------------+------------------+
| 总长度(4字节) | 类型长度(2字节) | 消息类型(N字节) | 请求ID(8字节) | 负载数据(M字节) | 
+------------------+------------------+------------------+------------------+------------------+

```
### 字段说明

1. **总长度** (4字节, big-endian): 整个消息的字节长度，包括头部和负载数据
2. **类型长度** (2字节, big-endian): 消息类型的字节长度
3. **消息类型** (N字节): UTF-8编码的字符串，表示消息类型，最大长度255字节
4. **请求ID** (8字节, big-endian): 用于标识请求-响应关系的唯一ID，推送消息时为0
5. **负载数据** (M字节): 经过序列化的消息内容

### 协议特点

- 使用大端序(big-endian)进行数字编码
- 消息类型必须是有效的UTF-8字符串，不包含控制字符
- 请求ID用于匹配请求和响应，当ID为0时表示推送消息
- 负载数据使用配置的序列化器进行序列化/反序列化

## 支持的协议

### TCP

chilix-msg 原生支持 TCP 协议，这是目前默认和主要的传输协议。TCP 提供了可靠的、面向连接的通信，适用于大多数应用场景。

### WebSocket (计划中)

WebSocket 支持正在开发中，将提供更好的浏览器兼容性和HTTP穿越能力。

### 自定义协议
```go

type CustomTransport struct {
	// 实现Transport 接口
}
func (t *CustomTransport) Listen(address string) (transport.Listener, error) { 
	// 实现监听逻辑 
}
func (t *CustomTransport) Dial(address string) (transport.Connection, error) { 
	// 实现拨号逻辑
}
```

## 核心概念

### Processor（处理器）

Processor 是 chilix-msg 的核心组件，负责处理网络连接上的消息。它提供了以下主要功能：

- 消息编解码
- 消息路由分发
- 请求-响应模式
- 消息推送

创建 Processor 的示例：
```go
// 创建处理器
processor := core.NewProcessor(conn, core.ProcessorOptions{
    Serializer:       serializer.DefaultSerializer,
    MessageSizeLimit: 1024 * 1024,
    RequestTimeout:   10 * time.Second,
})

```

### 消息处理

注册消息处理器来处理特定类型的消息：

```go
rocessor.RegisterHandler("message_type", func (ctx core.Context) error { 
	var payload MyPayloadType 
	if err := ctx.Bind(&payload); err != nil { 
		return err
	}
// 处理消息
// ...

// 可选地回复消息
return ctx.Reply(responseData)
})

```

### 请求-响应模式
chilix-msg 支持同步的请求-响应模式：
```go
// 发送请求并等待响应 
response, err := processor.Request("get_user", map[string]interface{}{ "user_id": 123, })
if err != nil {
    // 处理错误 
    return err
    }
var user User
if err := response.Bind(&user); err != nil {
    // 处理解析错误 
    return err
}

```

### 消息推送

服务器可以主动向客户端推送消息：
```go
// 服务器端推送消息 
err := processor.Send("notification", map[string]interface{}{ 
	"message": "Hello from server",
	"time": time.Now(), 
})
```
客户端注册相应的处理器来接收推送消息：
```go
processor.RegisterHandler("notification", func(ctx core.Context) error { var notification map[string]interface{} if err := ctx.Bind(&notification); err != nil { return err }
    log.Printf("Received notification: %v", notification)
    return nil
})

```

## 中间件支持

chilix-msg 支持中间件来增强消息处理功能：

```go
// 日志中间件示例 
func LoggingMiddleware() core.Middleware { 
	return func(next core.Handler) core.Handler { 
		return func(ctx core.Context) error { 
			log.Printf("Processing message: %s", ctx.MessageType()) 
			err := next(ctx) 
			log.Printf("Finished processing message: %s", ctx.MessageType()) 
			return err 
		} 
	} 
}
// 注册中间件 
processor.Use(LoggingMiddleware())
```

## 序列化
chilix-msg 默认使用 JSON 序列化，但您可以轻松替换为其他序列化方式：
```go
processor := core.NewProcessor(conn, core.ProcessorOptions{ 
	Serializer: serializer.DefaultSerializer,
})
```

这将启动一个服务器和一个客户端，演示请求-响应和消息推送功能。

## API 参考

### core.Processor

- `NewProcessor(conn transport.Connection, opts ProcessorOptions) *Processor` - 创建新的处理器
- [Use(middleware Middleware)](file:///Users/wry/IdeaProjects/chilix-msg/pkg/core/processor.go#L50-L52) - 注册中间件
- `RegisterHandler(msgType string, handler Handler)` - 注册消息处理器
- [Listen() error](file:///Users/wry/IdeaProjects/chilix-msg/pkg/core/processor.go#L61-L114) - 开始监听和处理消息
- `Send(msgType string, payload interface{}) error` - 发送消息
- `Request(msgType string, payload interface{}) (Response, error)` - 发送请求并等待响应
- `Reply(requestID uint64, msgType string, payload interface{}) error` - 发送响应
- [Close() error](file:///Users/wry/IdeaProjects/chilix-msg/pkg/core/processor.go#L158-L161) - 关闭处理器

## 贡献

欢迎提交 Issue 和 Pull Request 来改进 chilix-msg。

## 许可证

MIT
