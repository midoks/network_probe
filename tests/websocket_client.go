package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	serverURL = "ws://127.0.0.1:8082/ws"
)

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WebSocketResponse struct {
	Type    string      `json:"type"`
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// 从环境变量获取认证信息
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "xxx"
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "xxx"
	}

	// 创建 HTTP 请求头
	headers := http.Header{}
	headers.Set("X-Node-ID", nodeID)
	headers.Set("X-Secret", secret)

	// 连接 WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, headers)
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket")
	fmt.Printf("Using Node ID: %s\n", nodeID)

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 接收欢迎消息
	var welcome WebSocketResponse
	if err := conn.ReadJSON(&welcome); err != nil {
		log.Fatalf("Failed to read welcome message: %v", err)
	}
	fmt.Printf("Server: %+v\n\n", welcome)

	// 测试各种功能
	tests := []struct {
		name string
		fn   func(*websocket.Conn)
	}{
		{"MTR", testMTR},
		{"Ping", testPing},
		{"TCPing", testTCPing},
		{"Website", testWebsite},
		{"DNS", testDNS},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Testing %s ===\n", test.name)
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		test.fn(conn)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n=== All tests completed ===")
}

func testMTR(conn *websocket.Conn) {
	request := map[string]interface{}{
		"type": "mtr",
		"payload": map[string]interface{}{
			"host":     "baidu.com",
			"max_hops": 5,
			"count":    3,
			"interval": 1,
		},
	}

	sendAndReceive(conn, request)
}

func testPing(conn *websocket.Conn) {
	request := map[string]interface{}{
		"type": "ping",
		"payload": map[string]interface{}{
			"host":    "baidu.com",
			"count":   2,
			"timeout": 2,
		},
	}

	sendAndReceive(conn, request)
}

func testTCPing(conn *websocket.Conn) {
	request := map[string]interface{}{
		"type": "tcping",
		"payload": map[string]interface{}{
			"host":    "baidu.com",
			"port":    80,
			"count":   2,
			"timeout": 3,
		},
	}

	sendAndReceive(conn, request)
}

func testWebsite(conn *websocket.Conn) {
	request := map[string]interface{}{
		"type": "website",
		"payload": map[string]interface{}{
			"url": "https://www.baidu.com",
		},
	}

	sendAndReceive(conn, request)
}

func testDNS(conn *websocket.Conn) {
	request := map[string]interface{}{
		"type": "dns",
		"payload": map[string]interface{}{
			"domain":     "baidu.com",
			"query_type": "A",
		},
	}

	sendAndReceive(conn, request)
}

func sendAndReceive(conn *websocket.Conn, request map[string]interface{}) {
	// 发送请求
	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return
	}
	fmt.Printf("Sending:\n%s\n", jsonData)

	if err := conn.WriteJSON(request); err != nil {
		log.Printf("Failed to send request: %v", err)
		return
	}

	// 接收响应
	var response WebSocketResponse
	if err := conn.ReadJSON(&response); err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}

	// 格式化输出响应
	responseData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}
	fmt.Printf("Response:\n%s\n", responseData)

	// 检查响应状态
	if response.Status == "error" {
		fmt.Printf("⚠️  Error: %s\n", response.Message)
	} else {
		fmt.Printf("✅ Success\n")
	}
}
