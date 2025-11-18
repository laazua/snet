package main

import (
	"fmt"
	"log"
	"time"

	"github.com/laazua/snet/v2"
)

func main() {}

// 定义业务数据结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

// 登录处理器
type LoginHandler struct {
	v2.BaseHandler
}

func (h *LoginHandler) Handle(request v2.Request) {
	// 反序列化数据
	var loginReq LoginRequest
	err := request.GetConnection().(*snet.ConnectionImpl).Server.serializer.Deserialize(
		request.GetMessage().GetData(), &loginReq)
	if err != nil {
		log.Printf("Deserialize error: %v", err)
		return
	}

	fmt.Printf("Received login request: %+v\n", loginReq)

	// 处理业务逻辑
	response := LoginResponse{
		Code:    0,
		Message: "Login successful",
		UserID:  1001,
	}

	// 发送响应
	request.GetConnection().SendMsgWithStruct(2, response)
}

func startServer() {
	config := &snet.Config{
		Name:          "TestServer",
		Host:          "127.0.0.1",
		Port:          8888,
		WorkerNum:     100,
		MaxWorkerTask: 10000,
		MaxConn:       10000,
		UseTLS:        false, // 设置为true启用TLS
		CertFile:      "cert.pem",
		KeyFile:       "key.pem",
	}

	server := snet.NewServer(config)

	// 添加路由
	loginHandler := &LoginHandler{}
	server.AddRouter(1, loginHandler)

	// 启动服务器
	go server.Start()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
}
