package transport

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestKCPTransport_Basic(t *testing.T) {
	t.Parallel() // 启用并行测试

	tr := NewKCPTransport()
	addr := "127.0.0.1:0"

	// 启动监听器
	listener, err := tr.Listen(addr)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer listener.Close()

	realAddr := listener.Addr().String()
	fmt.Println("KCP real listen address:", realAddr)

	// 使用 WaitGroup 确保服务端启动
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 模拟客户端连接
		conn, err := tr.Dial(realAddr)
		if err != nil {
			t.Errorf("Dial failed: %v", err)
			return
		}
		defer conn.Close()

		msg := []byte("hello kcp")
		// 发送消息
		if _, err := conn.Write(msg); err != nil {
			t.Errorf("Write failed: %v", err)
			return
		}

		// 关键修复：等待服务端确认（避免连接过早关闭）
		time.Sleep(100 * time.Millisecond)
	}()

	// 等待服务端准备就绪（避免客户端过早连接）
	time.Sleep(50 * time.Millisecond)

	// 接受连接（带超时）
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	defer conn.Close()

	// 读取消息
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if n == 0 {
		t.Fatal("Read 0 bytes, expected message")
	}

	// 验证消息
	if string(buf[:n]) != "hello kcp" {
		t.Fatalf("Expected 'hello kcp', got %q", buf[:n])
	}

	// 等待客户端完成
	wg.Wait()
}
