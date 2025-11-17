package main

import (
	"fmt"
	"log"
	"time"

	"github.com/laazua/snet"
)

// 定义消息类型
const (
	MsgLogin = 1
	MsgChat  = 2
)

// 定义数据结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ChatMessage struct {
	From    string `json:"from"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
}

func main() {
	// 启动服务器
	go startServer()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 启动客户端
	startClient()
}

func startServer() {
	// 创建服务器配置
	config := snet.DefaultConfig()
	config.Address = ":8080"

	// 创建服务器
	server := snet.NewServer(config)

	// 注册处理器
	server.RegisterHandler(MsgLogin, handleLogin)
	server.RegisterHandler(MsgChat, handleChat)

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatal("Server start failed:", err)
	}
}

func handleLogin(ctx *snet.Context) error {
	var req LoginRequest
	if err := ctx.Request.ParseData(&req); err != nil {
		return err
	}

	// 处理登录逻辑
	response := LoginResponse{
		Success: req.Username == "admin" && req.Password == "123456",
		Message: "Login processed",
	}

	// 发送响应
	return ctx.Conn.SendMessage(ctx.Request.ID, response)
}

func handleChat(ctx *snet.Context) error {
	var msg ChatMessage
	if err := ctx.Request.ParseData(&msg); err != nil {
		return err
	}

	fmt.Printf("Chat message from %s: %s\n", msg.From, msg.Content)

	// 广播消息等逻辑...
	return nil
}

func startClient() {
	// 创建客户端
	client := snet.NewClient("localhost:8080", nil)

	// 连接服务器
	if err := client.Connect(); err != nil {
		log.Fatal("Client connect failed:", err)
	}
	defer client.Close()

	// 发送登录请求
	loginReq := LoginRequest{
		Username: "admin",
		Password: "123456",
	}

	// 发送并等待响应
	response, err := client.SendWithResponse(MsgLogin, loginReq, 5*time.Second)
	if err != nil {
		log.Fatal("Send login failed:", err)
	}

	var loginResp LoginResponse
	if err := response.ParseData(&loginResp); err != nil {
		log.Fatal("Parse response failed:", err)
	}

	fmt.Printf("Login response: %+v\n", loginResp)

	// 发送聊天消息
	chatMsg := ChatMessage{
		From:    "admin",
		Content: "Hello, World!",
		Time:    time.Now().Unix(),
	}

	if err := client.Send(MsgChat, chatMsg); err != nil {
		log.Fatal("Send chat failed:", err)
	}

	fmt.Println("Chat message sent successfully")
}
