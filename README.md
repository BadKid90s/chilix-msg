# 🚀 chilix-msg

<div align="center">

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![Go Report Card](https://goreportcard.com/badge/github.com/BadKid90s/chilix-msg)](https://goreportcard.com/report/github.com/BadKid90s/chilix-msg)
[![Coverage Status](https://img.shields.io/badge/coverage-90%25-brightgreen.svg)](#)
[![GitHub release](https://img.shields.io/github/release/BadKid90s/chilix-msg.svg)](https://GitHub.com/BadKid90s/chilix-msg/releases/)
[![GitHub tag](https://img.shields.io/github/tag/BadKid90s/chilix-msg.svg)](https://GitHub.com/BadKid90s/chilix-msg/tags/)
[![GitHub issues](https://img.shields.io/github/issues/BadKid90s/chilix-msg.svg)](https://GitHub.com/BadKid90s/chilix-msg/issues/)
[![GitHub pull-requests](https://img.shields.io/github/issues-pr/BadKid90s/chilix-msg.svg)](https://GitHub.com/BadKid90s/chilix-msg/pull/)
[![GitHub contributors](https://img.shields.io/github/contributors/BadKid90s/chilix-msg.svg)](https://GitHub.com/BadKid90s/chilix-msg/graphs/contributors/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/BadKid90s/chilix-msg/graphs/commit-activity)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/BadKid90s/chilix-msg.svg)](https://github.com/BadKid90s/chilix-msg)
[![GoDoc](https://godoc.org/github.com/BadKid90s/chilix-msg?status.svg)](https://godoc.org/github.com/BadKid90s/chilix-msg)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go)

<p></p>

[![GitHub stars](https://img.shields.io/github/stars/BadKid90s/chilix-msg?style=social)](https://github.com/BadKid90s/chilix-msg/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/BadKid90s/chilix-msg?style=social)](https://github.com/BadKid90s/chilix-msg/network)
[![GitHub watchers](https://img.shields.io/github/watchers/BadKid90s/chilix-msg?style=social)](https://github.com/BadKid90s/chilix-msg/watchers)

**高性能、轻量级的 Go 消息处理框架**

*专为网络 Socket 编程设计，提供简洁的 API 处理消息通信*

[特性](#-特性) • [安装](#-安装) • [快速开始](#-快速开始) • [使用场景](#-使用场景) • [API 文档](#-api-参考)

</div>

---

## 📖 概述

**chilix-msg** 是一个专为高性能网络通信而设计的 Go 语言消息处理框架。它采用模块化架构，支持多种传输协议和使用模式，让开发者能够轻松构建可扩展的网络应用。

### 🎯 设计理念

- **🔧 框架纯粹性**: 专注于消息传输和路由，不干涉业务逻辑
- **🎨 用户自由度**: 完全自定义消息格式和错误处理方式  
- **⚡ 高性能优先**: 基于 goroutine 的并发模型，零拷贝优化
- **🧩 模块化设计**: 可插拔的组件，易于扩展和定制

## 🌟 特性

<div align="center">

[![TCP](https://img.shields.io/badge/TCP-✓-success.svg)](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)
[![WebSocket](https://img.shields.io/badge/WebSocket-✓-success.svg)](https://en.wikipedia.org/wiki/WebSocket)
[![KCP](https://img.shields.io/badge/KCP-✓-success.svg)](https://github.com/skywind3000/kcp)
[![QUIC](https://img.shields.io/badge/QUIC-✓-success.svg)](https://en.wikipedia.org/wiki/QUIC)
[![AES](https://img.shields.io/badge/AES-✓-success.svg)](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard)
[![RSA](https://img.shields.io/badge/RSA-✓-success.svg)](https://en.wikipedia.org/wiki/RSA_(cryptosystem))
[![JSON](https://img.shields.io/badge/JSON-✓-success.svg)](https://www.json.org/)
[![Binary](https://img.shields.io/badge/Binary-✓-success.svg)](#)
[![Goroutine](https://img.shields.io/badge/Goroutine-✓-success.svg)](https://go.dev/)
[![Middleware](https://img.shields.io/badge/Middleware-✓-success.svg)](#)
[![Request/Response](https://img.shields.io/badge/Request%2FResponse-✓-success.svg)](#)
[![Push](https://img.shields.io/badge/Push-✓-success.svg)](#)

</div>

<div align="center">

| 🚀 性能 | 🧩 架构 | 🔌 协议 | 🔒 安全 |
|---------|--------|--------|--------|
| goroutine 并发 | 模块化设计 | TCP/WebSocket/KCP | AES/RSA 加密 |
| 零拷贝优化 | 中间件支持 | 自定义协议 | 端到端加密 |
| 二进制协议 | 可插拔序列化 | QUIC 支持 | 密钥管理 |

</div>

### 🚀 **高性能消息处理**
- 基于 goroutine 的并发处理模型
- 高效的二进制协议编解码
- 零拷贝消息处理优化

### 🧩 **模块化设计**
- 可插拔的序列化器（JSON、Binary 等）
- 灵活的中间件机制
- 易于扩展的传输层接口

### 🔌 **多协议支持**
- ✅ **TCP** - 默认支持，可靠传输
- ✅ **WebSocket** - 浏览器兼容，HTTP 穿越
- ✅ **KCP** - 低延迟 UDP 协议
- ✅ **QUIC** - 下一代传输协议
- 🔧 **自定义** - 支持自定义传输协议

### 🔒 **安全特性**
- 🔐 **对称加密** - AES-GCM 高性能加密
- 🔑 **非对称加密** - RSA 密钥交换
- 🔄 **自动密钥管理** - 透明的加解密处理

### ⚙️ **丰富功能**
- 🔄 **请求-响应模式** - 同步通信，自动请求匹配
- 📡 **消息推送** - 服务器主动推送，实时分发
- 🔧 **中间件支持** - 日志、异常恢复、加密等
- 📊 **可观测性** - 详细日志、性能指标、错误追踪

---

## 📦 安装

<div align="center">

[![Go Modules](https://img.shields.io/badge/Go%20Modules-supported-blue.svg)](https://github.com/golang/go/wiki/Modules)
[![go get](https://img.shields.io/badge/go%20get-supported-blue.svg)](https://golang.org/cmd/go/)
[![Semantic Versioning](https://img.shields.io/badge/SemVer-2.0.0-blue.svg)](https://semver.org/)

</div>

```bash
# 使用 go mod 安装
go get github.com/BadKid90s/chilix-msg

# 或者在您的 go.mod 文件中添加
require github.com/BadKid90s/chilix-msg latest
```

## 📚 使用场景

### 🔄 请求-响应模式

适用于需要同步获取结果的场景，如 API 调用、数据查询等。

<details>
<summary>📝 点击查看请求-响应示例</summary>

#### 服务端
```
package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    fmt.Println("✅ 服务器启动在 :8080")
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleRequestResponse(conn)
    }
}

func handleRequestResponse(conn net.Conn) {
    defer conn.Close()
    
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:       serializer.DefaultSerializer,
        MessageSizeLimit: 1024 * 1024,
        RequestTimeout:   10 * time.Second,
    })
    defer processor.Close()
    
    // 注册用户查询处理器
    processor.RegisterHandler("get_user", func(ctx core.Context) error {
        var req struct {
            UserID int `json:"user_id"`
        }
        if err := ctx.Bind(&req); err != nil {
            return ctx.Reply(map[string]interface{}{
                "success": false,
                "error":   "无效的请求格式",
            })
        }
        
        // 模拟数据库查询
        if req.UserID == 404 {
            return ctx.Reply(map[string]interface{}{
                "success": false,
                "error":   "用户不存在",
                "code":    "USER_NOT_FOUND",
            })
        }
        
        // 返回用户信息
        return ctx.Reply(map[string]interface{}{
            "success": true,
            "data": map[string]interface{}{
                "id":   req.UserID,
                "name": fmt.Sprintf("用户_%d", req.UserID),
                "email": fmt.Sprintf("user%d@example.com", req.UserID),
            },
        })
    })
    
    processor.Listen()
}
```

#### 客户端
```
package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:     serializer.DefaultSerializer,
        RequestTimeout: 5 * time.Second,
    })
    defer processor.Close()
    
    // 发送用户查询请求
    response, err := processor.Request("get_user", map[string]interface{}{
        "user_id": 123,
    })
    if err != nil {
        log.Printf("请求失败: %v", err)
        return
    }
    
    var result map[string]interface{}
    if err := response.Bind(&result); err != nil {
        log.Printf("解析响应失败: %v", err)
        return
    }
    
    if success, _ := result["success"].(bool); success {
        fmt.Printf("✅ 用户信息: %+v\n", result["data"])
    } else {
        fmt.Printf("❌ 错误: %s\n", result["error"])
    }
}
```
</details>

### 📡 消息推送模式

适用于实时通知、状态更新、事件分发等场景。

<details>
<summary>📝 点击查看消息推送示例</summary>

#### 服务端（推送服务）
```
package main

import (
    "fmt"
    "log"
    "net"
    "sync"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

type NotificationServer struct {
    clients map[string]*core.Processor
    mutex   sync.RWMutex
}

func NewNotificationServer() *NotificationServer {
    return &NotificationServer{
        clients: make(map[string]*core.Processor),
    }
}

func (ns *NotificationServer) AddClient(clientID string, processor *core.Processor) {
    ns.mutex.Lock()
    ns.clients[clientID] = processor
    ns.mutex.Unlock()
    fmt.Printf("➕ 客户端 %s 连接\n", clientID)
}

func (ns *NotificationServer) RemoveClient(clientID string) {
    ns.mutex.Lock()
    delete(ns.clients, clientID)
    ns.mutex.Unlock()
    fmt.Printf("➖ 客户端 %s 断开连接\n", clientID)
}

func (ns *NotificationServer) Broadcast(msgType string, data interface{}) {
    ns.mutex.RLock()
    defer ns.mutex.RUnlock()
    
    for clientID, processor := range ns.clients {
        if err := processor.Send(msgType, data); err != nil {
            fmt.Printf("发送失败给 %s: %v\n", clientID, err)
        }
    }
}

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    server := NewNotificationServer()
    
    // 启动定时推送
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        counter := 0
        for range ticker.C {
            counter++
            server.Broadcast("system_notification", map[string]interface{}{
                "id":      counter,
                "message": fmt.Sprintf("系统通知 #%d", counter),
                "time":    time.Now().Format(time.RFC3339),
                "type":    "info",
            })
        }
    }()
    
    fmt.Println("✅ 推送服务器启动在 :8080")
    
    clientID := 1
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        
        go func(id int, c net.Conn) {
            defer c.Close()
            clientIDStr := fmt.Sprintf("client_%d", id)
            
            processor := core.NewProcessor(c, core.ProcessorOptions{
                Serializer: serializer.DefaultSerializer,
            })
            defer processor.Close()
            
            server.AddClient(clientIDStr, processor)
            defer server.RemoveClient(clientIDStr)
            
            // 注册客户端上线通知
            processor.RegisterHandler("client_online", func(ctx core.Context) error {
                var req struct {
                    Username string `json:"username"`
                }
                ctx.Bind(&req)
                
                // 广播用户上线消息
                server.Broadcast("user_online", map[string]interface{}{
                    "username": req.Username,
                    "time":     time.Now().Format(time.RFC3339),
                })
                
                return nil
            })
            
            processor.Listen()
        }(clientID, conn)
        
        clientID++
    }
}
```

#### 客户端（接收推送）
```
package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer: serializer.DefaultSerializer,
    })
    defer processor.Close()
    
    // 注册系统通知处理器
    processor.RegisterHandler("system_notification", func(ctx core.Context) error {
        var notification map[string]interface{}
        if err := ctx.Bind(&notification); err != nil {
            return err
        }
        
        fmt.Printf("🔔 %s: %s\n", 
            notification["time"], 
            notification["message"])
        return nil
    })
    
    // 注册用户上线通知处理器
    processor.RegisterHandler("user_online", func(ctx core.Context) error {
        var event map[string]interface{}
        if err := ctx.Bind(&event); err != nil {
            return err
        }
        
        fmt.Printf("🟢 用户 %s 上线了\n", event["username"])
        return nil
    })
    
    // 启动监听
    go func() {
        if err := processor.Listen(); err != nil {
            log.Printf("监听错误: %v", err)
        }
    }()
    
    // 发送上线通知
    time.Sleep(100 * time.Millisecond)
    processor.Send("client_online", map[string]interface{}{
        "username": "Alice",
    })
    
    // 保持连接
    fmt.Println("✅ 已连接到推送服务器，等待通知...")
    select {}
}
```
</details>

### 🔀 混合模式

在实际应用中，通常需要同时支持请求-响应和消息推送。

<details>
<summary>📝 点击查看混合模式示例</summary>

```go
package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    // 客户端示例：同时处理请求-响应和推送消息
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:     serializer.DefaultSerializer,
        RequestTimeout: 5 * time.Second,
    })
    defer processor.Close()
    
    // 注册推送消息处理器
    processor.RegisterHandler("notification", func(ctx core.Context) error {
        var msg map[string]interface{}
        if err := ctx.Bind(&msg); err != nil {
            return err
        }
        fmt.Printf("🔔 收到推送: %s\n", msg["content"])
        return nil
    })
    
    // 启动监听推送消息
    go func() {
        processor.Listen()
    }()
    
    time.Sleep(100 * time.Millisecond)
    
    // 发送请求获取数据
    response, err := processor.Request("get_user", map[string]interface{}{
        "user_id": 123,
    })
    if err != nil {
        log.Printf("请求失败: %v", err)
        return
    }
    
    var result map[string]interface{}
    response.Bind(&result)
    fmt.Printf("📊 请求响应: %+v\n", result)
    
    // 继续接收推送消息
    time.Sleep(10 * time.Second)
}
```
</details>

---

## 🚀 快速开始

### 服务端

```go
package main

import (
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal("服务器启动失败:", err)
    }
    defer listener.Close()
    log.Println("✅ 服务器启动在 :8080")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Println("接受连接失败:", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    // 创建处理器
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:       serializer.DefaultSerializer,
        MessageSizeLimit: 1024 * 1024,
        RequestTimeout:   10 * time.Second,
    })
    defer processor.Close()

    // 注册消息处理器
    processor.RegisterHandler("get_time", func(ctx core.Context) error {
        currentTime := time.Now().Format(time.RFC3339)
        return ctx.Reply(currentTime)
    })

    // 启动监听
    if err := processor.Listen(); err != nil {
        log.Printf("连接错误: %v", err)
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
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer conn.Close()
    
    // 创建处理器
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:       serializer.DefaultSerializer,
        MessageSizeLimit: 1024 * 1024,
        RequestTimeout:   10 * time.Second,
    })
    defer processor.Close()

    // 发送时间请求
    response, err := processor.Request("get_time", nil)
    if err != nil {
        log.Printf("时间请求失败: %v", err)
    } else {
        var timeStr string
        if err := response.Bind(&timeStr); err != nil {
            log.Printf("解析时间响应失败: %v", err)
        } else {
            log.Printf("⏰ 服务器时间: %s", timeStr)
        }
    }
}
```

## 📚 协议格式

chilix-msg 使用基于长度前缀的二进制协议格式，确保消息的可靠传输和解析。

### 消息结构

```
+--------------+----------------+----------------+--------------+----------------+
| 总长度(4字节) | 类型长度(2字节) | 消息类型(N字节) | 请求ID(8字节) | 负载数据(M字节) | 
+--------------+----------------+----------------+--------------+----------------+
```

### 字段说明

| 字段 | 长度 | 编码 | 说明 |
|------|------|------|------|
| 总长度 | 4字节 | big-endian | 整个消息的字节长度，包括头部和负载数据 |
| 类型长度 | 2字节 | big-endian | 消息类型的字节长度 |
| 消息类型 | N字节 | UTF-8 | 表示消息类型，最大长度255字节 |
| 请求ID | 8字节 | big-endian | 用于标识请求-响应关系的唯一ID，推送消息时为0 |
| 负载数据 | M字节 | 序列化 | 经过序列化的消息内容 |

### 协议特点

- **大端序**: 使用大端序(big-endian)进行数字编码
- **UTF-8 支持**: 消息类型必须是有效的UTF-8字符串
- **请求匹配**: 请求ID用于匹配请求和响应，为0时表示推送消息
- **灵活序列化**: 负载数据使用配置的序列化器进行序列化/反序列化

---
## 🔌 支持的协议

<div align="center">

| 协议 | 状态 | 特点 | 适用场景 |
|------|------|------|----------|
| **TCP** | ✅ 支持 | 可靠传输、面向连接 | 大部分应用场景 |
| **WebSocket** | ✅ 支持 | 浏览器兼容、HTTP 穿越 | Web 应用、实时通信 |
| **KCP** | ✅ 支持 | 低延迟、快速重传 | 游戏、音视频通信 |
| **QUIC** | ✅ 支持 | 低延迟、多路复用 | 下一代网络应用 |
| **自定义** | 🔧 扩展 | 灵活定制 | 特殊需求 |

</div>

### 自定义协议示例

```
type CustomTransport struct {
    // 实现 Transport 接口
}

func (t *CustomTransport) Listen(address string) (transport.Listener, error) {
    // 实现监听逻辑
}

func (t *CustomTransport) Dial(address string) (transport.Connection, error) {
    // 实现拨号逻辑
}
```

---

## 📚 核心概念

### 📦 Processor（处理器）

**Processor** 是 chilix-msg 的核心组件，负责处理网络连接上的消息。

#### 主要功能：
- 🔄 消息编解码和路由分发
- 📡 请求-响应模式和消息推送
- 🔧 中间件支持和扩展
- ⚙️ 配置管理和生命周期控制

#### 创建示例：
```go
processor := core.NewProcessor(conn, core.ProcessorOptions{
    Serializer:       serializer.DefaultSerializer,
    MessageSizeLimit: 1024 * 1024,
    RequestTimeout:   10 * time.Second,
    Logger:           log.NewDefaultLogger(),
})
```

### 📋 Context（上下文）

**Context** 提供了完整的消息处理上下文，包含消息信息和响应方法。

#### 常用方法：
```go
processor.RegisterHandler("message_type", func(ctx core.Context) error {
    // 获取消息信息
    msgType := ctx.MessageType()    // 消息类型
    requestID := ctx.RequestID()    // 请求ID
    
    // 绑定消息负载
    var payload MyPayloadType
    if err := ctx.Bind(&payload); err != nil {
        return err
    }
    
    // 发送响应（只限请求消息）
    if ctx.IsRequest() {
        return ctx.Reply(responseData)
    }
    
    return nil
})
```

### 📨 Writer（写入器）

**Writer** 提供了消息发送接口，支持不同的发送模式。

#### 方法说明：
```go
// 发送普通消息（推送）
writer.Write("notification", data)

// 发送响应消息
writer.Reply(requestID, "response_type", data)
```

### 💬 Response（响应）

**Response** 封装了请求的响应结果，提供数据绑定和信息获取功能。

#### 使用示例：
```go
response, err := processor.Request("get_user", userRequest)
if err != nil {
    // 处理通信错误
    return err
}

// 绑定响应数据
var user User
if err := response.Bind(&user); err != nil {
    // 处理数据解析错误
    return err
}

// 检查响应信息
fmt.Printf("响应类型: %s\n", response.MsgType())
fmt.Printf("请求ID: %d\n", response.RequestID())
```

---

## 🔧 中间件支持

chilix-msg 提供强大的中间件系统，允许您轻松扩展消息处理功能。

### 📝 日志中间件

```go
func LoggingMiddleware() core.Middleware {
    return func(next core.Handler) core.Handler {
        return func(ctx core.Context) error {
            start := time.Now()
            
            log.Printf("⬇️  处理消息: %s [%d]", 
                ctx.MessageType(), ctx.RequestID())
            
            err := next(ctx)
            
            log.Printf("⬆️  完成处理: %s [%d] 耗时: %v", 
                ctx.MessageType(), ctx.RequestID(), time.Since(start))
            
            return err
        }
    }
}

// 注册中间件
processor.Use(LoggingMiddleware())
```

### 🔒 加密中间件

#### AES 对称加密
```go
import "github.com/BadKid90s/chilix-msg/middleware"

// 生成加密密钥
encryptionKey := middleware.KeyFromString("您的密钥")

// 在客户端和服务端都注册加密中间件
processor.Use(middleware.EncryptionMiddleware(encryptionKey))
```

#### RSA 非对称加密
```go
// 生成RSA密钥对
privateKey, publicKey, err := middleware.GenerateRSAKeyPair(2048)
if err != nil {
    log.Fatal("生成RSA密钥对失败:", err)
}

// 注册加密中间件
processor.Use(middleware.RSAEncryptionMiddleware(privateKey, publicKey))
```

### ⚙️ 自定义中间件

```go
func AuthenticationMiddleware(secretKey string) core.Middleware {
    return func(next core.Handler) core.Handler {
        return func(ctx core.Context) error {
            // 检查认证信息
            if !isAuthenticated(ctx, secretKey) {
                // 认证失败是框架层错误，直接返回error
                return fmt.Errorf("authentication failed")
            }
            
            // 继续处理
            return next(ctx)
        }
    }
}

// 注册中间件
processor.Use(AuthenticationMiddleware("秘密密钥"))
```

### 🔄 中间件链

中间件按照注册顺序执行：

```go
// 执行顺序：日志 -> 认证 -> 加密 -> 处理器
processor.Use(LoggingMiddleware())
processor.Use(AuthenticationMiddleware("secret"))
processor.Use(middleware.EncryptionMiddleware(key))
```

---

## 📚 序列化

chilix-msg 默认使用 JSON 序列化，但您可以轻松替换为其他序列化方式：

### 默认序列化器
```go
processor := core.NewProcessor(conn, core.ProcessorOptions{
    Serializer: serializer.DefaultSerializer, // JSON 序列化
})
```

### 自定义序列化器
```go
// 使用 Binary 序列化
processor := core.ProcessorOptions{
    Serializer: &serializer.Binary{},
}

// 或者实现自定义序列化器
type CustomSerializer struct{}

func (s *CustomSerializer) Serialize(data interface{}) ([]byte, error) {
    // 实现序列化逻辑
}

func (s *CustomSerializer) Deserialize(data []byte, target interface{}) error {
    // 实现反序列化逻辑
}
```

---

## 📈 性能优化

### 配置优化

```go
processor := core.NewProcessor(conn, core.ProcessorOptions{
    Serializer:       serializer.DefaultSerializer,
    MessageSizeLimit: 10 * 1024 * 1024,    // 10MB 消息大小限制
    RequestTimeout:   30 * time.Second,     // 30秒请求超时
    Logger:           log.NewDefaultLogger(),
})
```

### 并发处理

```go
// 服务端并发处理多个连接
for {
    conn, err := listener.Accept()
    if err != nil {
        continue
    }
    
    // 每个连接在独立的 goroutine 中处理
    go handleConnection(conn)
}
```

---

## 🔒 加密通信

chilix-msg 提供了强大的加密通信功能，支持对称加密和非对称加密两种方式，确保数据在网络传输过程中的安全。

### 🔐 对称加密 (AES-GCM)

对称加密使用相同的密钥进行加密和解密，具有高性能的特点，适合大量数据的加密传输。

<details>
<summary>📝 点击查看对称加密示例</summary>

#### 服务端
```
package main

import (
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/middleware"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    // 生成加密密钥
    encryptionKey := middleware.KeyFromString("my-secret-password")
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        
        go func(c net.Conn) {
            defer c.Close()
            
            processor := core.NewProcessor(c, core.ProcessorOptions{
                Serializer: serializer.DefaultSerializer,
            })
            defer processor.Close()
            
            // 注册加密中间件
            processor.Use(middleware.EncryptionMiddleware(encryptionKey))
            
            processor.RegisterHandler("secure_message", func(ctx core.Context) error {
                var msg map[string]interface{}
                if err := ctx.Bind(&msg); err != nil {
                    return err
                }
                
                log.Printf("🔓 解密消息: %+v", msg)
                
                return ctx.Reply(map[string]interface{}{
                    "status": "received",
                    "echo":   msg,
                })
            })
            
            processor.Listen()
        }(conn)
    }
}
```

#### 客户端
```
package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/middleware"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    processor := core.NewProcessor(conn, core.ProcessorOptions{
        Serializer:     serializer.DefaultSerializer,
        RequestTimeout: 5 * time.Second,
    })
    defer processor.Close()
    
    // 使用相同的密钥
    encryptionKey := middleware.KeyFromString("my-secret-password")
    processor.Use(middleware.EncryptionMiddleware(encryptionKey))
    
    // 发送加密消息
    response, err := processor.Request("secure_message", map[string]interface{}{
        "sensitive_data": "这是机密信息",
        "user_id":        12345,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    var result map[string]interface{}
    response.Bind(&result)
    fmt.Printf("🔒 加密通信成功: %+v\n", result)
}
```
</details>

#### 密钥管理
```
// 从字符串生成密钥（推荐方式）
key1 := middleware.KeyFromString("my-secret-password")

// 从Base64编码的字符串生成密钥
key2, err := middleware.KeyFromBase64("base64-encoded-key")

// 直接使用字节密钥（必须是16、24或32字节）
key3 := []byte("1234567890123456") // 16字节AES-128密钥
```

### 🔑 非对称加密 (RSA)

非对称加密使用公钥加密、私钥解密，提供了更高的安全性，特别适合密钥分发和身份验证场景。

<details>
<summary>📝 点击查看非对称加密示例</summary>

```go
package main

import (
    "crypto/rsa"
    "log"
    "net"
    "github.com/BadKid90s/chilix-msg/core"
    "github.com/BadKid90s/chilix-msg/middleware"
    "github.com/BadKid90s/chilix-msg/serializer"
)

func main() {
    // 生成RSA密钥对（通常在服务端完成）
    privateKey, publicKey, err := middleware.GenerateRSAKeyPair(2048)
    if err != nil {
        log.Fatal("生成RSA密钥对失败:", err)
    }
    
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()
    
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        
        go func(c net.Conn) {
            defer c.Close()
            
            processor := core.NewProcessor(c, core.ProcessorOptions{
                Serializer: serializer.DefaultSerializer,
            })
            defer processor.Close()
            
            // 注册RSA加密中间件
            processor.Use(middleware.RSAEncryptionMiddleware(privateKey, publicKey))
            
            processor.RegisterHandler("rsa_message", func(ctx core.Context) error {
                var msg map[string]interface{}
                if err := ctx.Bind(&msg); err != nil {
                    return err
                }
                
                log.Printf("🔓 RSA解密消息: %+v", msg)
                return ctx.Reply(map[string]string{"status": "success"})
            })
            
            processor.Listen()
        }(conn)
    }
}
```
</details>

#### 密钥管理
```
// 生成RSA密钥对
privateKey, publicKey, err := middleware.GenerateRSAKeyPair(2048) // 支持1024、2048、4096位

// 导出密钥为PEM格式
privateKeyPEM := middleware.ExportRSAPrivateKey(privateKey)
publicKeyPEM := middleware.ExportRSAPublicKey(publicKey)

// 从PEM格式导入密钥
importedPrivateKey, err := middleware.LoadRSAPrivateKey(privateKeyPEM)
importedPublicKey, err := middleware.LoadRSAPublicKey(publicKeyPEM)
```

### 🛡️ 加密机制说明

<div align="center">

| 特性 | AES-GCM | RSA-OAEP |
|------|---------|----------|
| **算法类型** | 对称加密 | 非对称加密 |
| **密钥长度** | 128/192/256 位 | 1024/2048/4096 位 |
| **性能** | 高性能 | 相对较慢 |
| **适用场景** | 大量数据传输 | 密钥交换、身份验证 |
| **安全特性** | 认证加密 | 数字签名 |

</div>

#### 对称加密 (AES-GCM)
- ✅ 使用 AES-GCM 算法提供认证加密
- ✅ 支持 128、192、256 位密钥长度
- ✅ 自动处理 nonce 生成
- ✅ 提供数据完整性和机密性保护

#### 非对称加密 (RSA)
- ✅ 使用 RSA-OAEP 算法进行密钥加密
- ✅ 采用混合加密模式：RSA 加密 AES 密钥，AES 加密实际数据
- ✅ 支持 1024、2048、4096 位密钥长度
- ✅ 提供身份验证和密钥分发能力

### 🔐 安全建议

#### 密钥管理
- **对称加密**: 使用强密码生成密钥，定期更换
- **非对称加密**: 保护好私钥，公钥可以公开分发

#### 密钥分发
- **对称加密**: 需要安全的密钥分发机制
- **非对称加密**: 可以通过安全渠道分发公钥

#### 性能考虑
- **对称加密**: 适合大量数据加密
- **非对称加密**: 适合密钥交换和身份验证，性能相对较低

#### 混合使用
- 可以结合使用两种加密方式，发挥各自优势
- 典型模式：RSA 交换 AES 密钥，AES 加密实际通信数据

---

## 📚 API 参考

### 🔧 core.Processor

**消息处理器的核心接口，负责网络连接上的消息处理。**

#### 创建和配置
```go
// 创建新的处理器
func NewProcessor(conn transport.Connection, opts ProcessorOptions) *Processor

// 处理器选项
type ProcessorOptions struct {
    Serializer       serializer.Serializer // 序列化器
    MessageSizeLimit int                    // 消息大小限制
    RequestTimeout   time.Duration          // 请求超时时间
    Logger           log.Logger             // 日志记录器
}
```

#### 中间件和处理器
```go
// 注册中间件
func (p *Processor) Use(middleware Middleware)

// 注册消息处理器
func (p *Processor) RegisterHandler(msgType string, handler Handler)

// 处理器函数签名
type Handler func(ctx Context) error

// 中间件函数签名
type Middleware func(next Handler) Handler
```

#### 消息通信
```go
// 开始监听和处理消息
func (p *Processor) Listen() error

// 发送消息（推送模式）
func (p *Processor) Send(msgType string, payload interface{}) error

// 发送请求并等待响应（请求-响应模式）
func (p *Processor) Request(msgType string, payload interface{}) (Response, error)

// 发送响应
func (p *Processor) Reply(requestID uint64, msgType string, payload interface{}) error

// 关闭处理器
func (p *Processor) Close() error
```

### 📋 core.Context

**消息处理上下文，提供完整的消息信息和响应方法。**

```go
type Context interface {
    // 消息信息
    MessageType() string              // 获取消息类型
    RequestID() uint64                // 获取请求ID
    IsRequest() bool                  // 判断是否是请求消息
    IsResponse() bool                 // 判断是否是响应消息
    RawData() []byte                  // 获取原始数据
    
    // 数据绑定
    Bind(target interface{}) error    // 绑定消息负载
    
    // 连接和组件
    Connection() transport.Connection // 获取底层连接
    Writer() Writer                   // 获取消息写入器
    Logger() log.Logger               // 获取日志记录器
    Processor() *Processor            // 获取处理器
    
    // 响应方法
    Reply(payload interface{}) error  // 发送成功响应
}
```

### 📨 core.Writer

**消息写入器接口，提供消息发送功能。**

```go
type Writer interface {
    // 发送消息（推送模式）
    Write(msgType string, payload interface{}) error
    
    // 发送响应（请求-响应模式）
    Reply(requestID uint64, msgType string, payload interface{}) error
}
```

### 💬 core.Response

**响应接口，封装请求的响应结果。**

```go
type Response interface {
    // 响应信息
    MsgType() string                  // 获取响应消息类型
    RequestID() uint64                // 获取请求ID
    RawData() []byte                  // 获取原始响应数据
    
    // 数据绑定
    Bind(target interface{}) error    // 绑定响应数据
}
```

### 🔌 transport.Transport

**传输层接口，支持多种网络协议。**

```go
type Transport interface {
    // 监听连接
    Listen(address string) (Listener, error)
    
    // 拨号连接
    Dial(address string) (Connection, error)
}

// 支持的传输协议
- TCP:       transport.NewTCPTransport()
- WebSocket: transport.NewWebSocketTransport()
- KCP:       transport.NewKCPTransport()
- QUIC:      transport.NewQUICTransport()
```

---

## 🔧 错误处理

chilix-msg 采用纯粹的框架设计，不干涉业务逻辑，所有错误处理完全由用户自定义。

### 错误类型

#### 🚈 通信错误
通过函数返回值 `error` 处理，表示网络传输、协议解析等框架层面的错误：

```go
// 请求发送失败
response, err := processor.Request("get_user", userData)
if err != nil {
    log.Printf("🚨 通信错误: %v", err)
    return
}

// 消息发送失败
if err := processor.Send("notification", data); err != nil {
    log.Printf("🚨 发送失败: %v", err)
}
```

#### 💼 业务错误
通过自定义响应结构处理，完全由用户控制格式和内容：

```go
// 服务端：返回业务错误
processor.RegisterHandler("get_user", func(ctx core.Context) error {
    var req struct {
        UserID int `json:"user_id"`
    }
    
    if err := ctx.Bind(&req); err != nil {
        // 返回自定义错误格式
        return ctx.Reply(map[string]interface{}{
            "success": false,
            "error":   "无效的请求格式",
            "code":    "INVALID_REQUEST",
        })
    }
    
    // 正常返回数据
    return ctx.Reply(map[string]interface{}{
        "success": true,
        "data":    getUserData(req.UserID),
    })
})

// 客户端：处理业务错误
response, err := processor.Request("get_user", map[string]int{"user_id": 123})
if err != nil {
    // 处理通信错误
    log.Printf("🚨 请求失败: %v", err)
    return
}

var result map[string]interface{}
if err := response.Bind(&result); err != nil {
    log.Printf("🚨 解析响应失败: %v", err)
    return
}

// 检查业务结果
if success, ok := result["success"].(bool); ok && success {
    fmt.Printf("😎 用户数据: %+v", result["data"])
} else {
    fmt.Printf("⚠️ 业务错误: %s", result["error"])
}
```

## 🧪 测试

### 单元测试

```go
func TestMessageHandler(t *testing.T) {
    // 创建测试连接
    server, client := net.Pipe()
    defer server.Close()
    defer client.Close()
    
    // 创建处理器
    processor := core.NewProcessor(server, core.ProcessorOptions{
        Serializer: serializer.DefaultSerializer,
    })
    defer processor.Close()
    
    // 注册测试处理器
    processor.RegisterHandler("test", func(ctx core.Context) error {
        return ctx.Reply("success")
    })
    
    // 启动处理器
    go processor.Listen()
    
    // 测试消息发送
    clientProcessor := core.NewProcessor(client, core.ProcessorOptions{
        Serializer: serializer.DefaultSerializer,
    })
    defer clientProcessor.Close()
    
    go clientProcessor.Listen()
    
    response, err := clientProcessor.Request("test", "hello")
    assert.NoError(t, err)
    
    var result string
    err = response.Bind(&result)
    assert.NoError(t, err)
    assert.Equal(t, "success", result)
}
```



## 🌍 社区

<div align="center">

### 💬 加入讨论

[![GitHub Discussions](https://img.shields.io/badge/GitHub-Discussions-green?logo=github)](https://github.com/BadKid90s/chilix-msg/discussions)
[![Stack Overflow](https://img.shields.io/badge/Stack-Overflow-orange?logo=stackoverflow)](https://stackoverflow.com/questions/tagged/chilix-msg)

### 🐛 反馈问题

发现问题？有功能建议？欢迎提交 [Issue](https://github.com/BadKid90s/chilix-msg/issues)

### 🔥 贡献代码

欢迎提交 [Pull Request](https://github.com/BadKid90s/chilix-msg/pulls) 来改进 chilix-msg！

</div>

---

## 🙏 致谢

感谢所有为 chilix-msg 做出贡献的开发者！

<div align="center">

[![Contributors](https://img.shields.io/github/contributors/BadKid90s/chilix-msg?style=for-the-badge)](https://github.com/BadKid90s/chilix-msg/graphs/contributors)

</div>

---

## 📋 许可证

<div align="center">

**MIT License**

本项目采用 [MIT 许可证](LICENSE)，您可以自由使用、修改和分发。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

</div>

---

<div align="center">

**✨ 如果 chilix-msg 对您有帮助，请给我们一个 Star ⭐ ✨**



*由 ❤️ 心制作，为开发者打造更好的工具*

</div>