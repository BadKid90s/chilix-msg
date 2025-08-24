package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	ServerAddress = "127.0.0.1:8080"
	BufferSize    = 1024
)

// 协议格式：4字节长度（大端序） + 消息内容
func main() {
	//if len(os.Args) < 2 {
	//	fmt.Println("Usage:")
	//	fmt.Println("  go run tcp_demo.go server - Start as server")
	//	fmt.Println("  go run tcp_demo.go client - Start as client")
	//	return
	//}
	//
	//switch os.Args[1] {
	//case "server":
	//	startServer()
	//case "client":
	//	startClient()
	//default:
	//	fmt.Println("Invalid command. Use 'server' or 'client'")
	//}

	go startServer()

	startClient()
}

// ====================== 服务器实现 ======================

func startServer() {
	listener, err := net.Listen("tcp", ServerAddress)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
	defer listener.Close()

	log.Printf("✅ Server started on %s", ServerAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept connection failed: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Client connected: %s", conn.RemoteAddr())

	// 启动主动发送数据的goroutine
	go sendPeriodicUpdates(conn)

	// 处理客户端请求
	for {
		// 读取消息长度
		lenBuf := make([]byte, 4)
		if _, err := conn.Read(lenBuf); err != nil {
			log.Printf("Failed to read message length: %v", err)
			return
		}

		// 解析消息长度
		msgLen := binary.BigEndian.Uint32(lenBuf)
		if msgLen > BufferSize {
			log.Printf("Message too large: %d > %d", msgLen, BufferSize)
			return
		}

		// 读取消息内容
		msgBuf := make([]byte, msgLen)
		if _, err := conn.Read(msgBuf); err != nil {
			log.Printf("Failed to read message: %v", err)
			return
		}

		message := string(msgBuf)
		log.Printf("Received: %s", message)

		// 处理不同类型的消息
		switch {
		case strings.HasPrefix(message, "ECHO:"):
			// Echo处理：原样返回消息
			echoMessage := strings.TrimPrefix(message, "ECHO:")
			sendMessage(conn, "ECHO_RESPONSE:"+echoMessage)

		case message == "GET_TIME":
			// 请求-响应模式：返回当前时间
			currentTime := time.Now().Format(time.RFC3339)
			sendMessage(conn, "TIME_RESPONSE:"+currentTime)

		case message == "EXIT":
			// 退出命令
			log.Printf("Client requested exit")
			return

		default:
			// 未知命令
			sendMessage(conn, "ERROR:Unknown command")
		}
	}
}

// 定期向客户端发送更新
func sendPeriodicUpdates(conn net.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			update := fmt.Sprintf("SERVER_UPDATE:%d", counter)
			if err := sendMessage(conn, update); err != nil {
				log.Printf("Failed to send update: %v", err)
				return
			}
		}
	}
}

// ====================== 客户端实现 ======================

func startClient() {
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	log.Printf("✅ Connected to server at %s", ServerAddress)

	// 启动接收消息的goroutine
	go receiveMessages(conn)

	// 读取用户输入
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter commands (ECHO:<text>, GET_TIME, EXIT):")

	for scanner.Scan() {
		command := scanner.Text()

		switch {
		case strings.HasPrefix(command, "ECHO:"):
			// 发送ECHO请求
			if err := sendMessage(conn, command); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}

		case command == "GET_TIME":
			// 发送时间请求
			if err := sendMessage(conn, command); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}

		case command == "EXIT":
			// 发送退出命令
			if err := sendMessage(conn, command); err != nil {
				log.Printf("Failed to send message: %v", err)
			}
			return

		default:
			fmt.Println("Unknown command. Valid commands: ECHO:<text>, GET_TIME, EXIT")
		}
	}
}

// 接收服务器消息
func receiveMessages(conn net.Conn) {
	for {
		// 读取消息长度
		lenBuf := make([]byte, 4)
		if _, err := conn.Read(lenBuf); err != nil {
			log.Printf("Failed to read message length: %v", err)
			return
		}

		// 解析消息长度
		msgLen := binary.BigEndian.Uint32(lenBuf)
		if msgLen > BufferSize {
			log.Printf("Message too large: %d > %d", msgLen, BufferSize)
			return
		}

		// 读取消息内容
		msgBuf := make([]byte, msgLen)
		if _, err := conn.Read(msgBuf); err != nil {
			log.Printf("Failed to read message: %v", err)
			return
		}

		message := string(msgBuf)

		// 处理不同类型的响应
		switch {
		case strings.HasPrefix(message, "ECHO_RESPONSE:"):
			response := strings.TrimPrefix(message, "ECHO_RESPONSE:")
			fmt.Printf("Server echoed: %s\n", response)

		case strings.HasPrefix(message, "TIME_RESPONSE:"):
			response := strings.TrimPrefix(message, "TIME_RESPONSE:")
			fmt.Printf("Server time: %s\n", response)

		case strings.HasPrefix(message, "SERVER_UPDATE:"):
			update := strings.TrimPrefix(message, "SERVER_UPDATE:")
			fmt.Printf("Server update: %s\n", update)

		case strings.HasPrefix(message, "ERROR:"):
			errorMsg := strings.TrimPrefix(message, "ERROR:")
			fmt.Printf("Error from server: %s\n", errorMsg)

		default:
			fmt.Printf("Received unknown message: %s\n", message)
		}
	}
}

// ====================== 通用工具函数 ======================

// sendMessage 发送消息（4字节长度 + 消息内容）
func sendMessage(conn net.Conn, message string) error {
	msgBytes := []byte(message)
	msgLen := uint32(len(msgBytes))

	// 创建缓冲区（4字节长度 + 消息内容）
	buf := make([]byte, 4+msgLen)

	// 写入消息长度（大端序）
	binary.BigEndian.PutUint32(buf[0:4], msgLen)

	// 写入消息内容
	copy(buf[4:], msgBytes)

	// 发送消息
	_, err := conn.Write(buf)
	return err
}

// receiveMessage 接收消息（先读4字节长度，再读消息内容）
func receiveMessage(conn net.Conn) (string, error) {
	// 读取消息长度
	lenBuf := make([]byte, 4)
	if _, err := conn.Read(lenBuf); err != nil {
		return "", err
	}

	// 解析消息长度
	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > BufferSize {
		return "", errors.New("message too large")
	}

	// 读取消息内容
	msgBuf := make([]byte, msgLen)
	if _, err := conn.Read(msgBuf); err != nil {
		return "", err
	}

	return string(msgBuf), nil
}
