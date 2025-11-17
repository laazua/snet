package main

import (
	"log"
	"time"

	"github.com/laazua/snet/v2"
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
	if err := ctx.Request.Unmarshal(&req); err != nil {
		return err
	}

	log.Printf("Login attempt: %s", req.Username)

	resp := LoginResponse{
		Success: true,
		Message: "Login successful",
	}

	return ctx.Conn.Send(ctx.Request.ID, resp)
}
