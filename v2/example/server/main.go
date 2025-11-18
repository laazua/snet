package main

import (
	"context"
	"log"

	v2 "github.com/laazua/snet/v2"
)

func main() {
	startServer()
}

// Request 请求结构体
type Request struct {
	Action string `json:"action"`
	Data   any    `json:"data"`
}

// Response 响应结构体
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// EchoHandler 回显处理器
type EchoHandler struct{}

func (e EchoHandler) Handle(ctx context.Context, req []byte) ([]byte, error) {
	var request Request
	codec := v2.NewDefaultCodec()

	if err := codec.Decode(req, &request); err != nil {
		return nil, err
	}

	response := Response{
		Code:    0,
		Message: "success",
		Data:    request.Data,
	}

	return codec.Encode(response)
}

func startServer() {
	config := &v2.ServerConfig{
		Addr:        ":8080",
		WorkerCount: 100,
		QueueSize:   1000,
	}

	server := v2.NewServer(config)
	server.RegisterHandler(EchoHandler{})

	if err := server.Start(); err != nil {
		log.Fatal("Server start failed:", err)
	}

	log.Println("Server started on :8080")

}
