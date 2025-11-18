package main

import (
	"log"
	"time"

	v2 "github.com/laazua/snet/v2"
)

func main() {
	startClient()
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

func startClient() {
	client := v2.NewClient(nil, v2.NewDefaultCodec())

	if err := client.Connect("localhost:8080"); err != nil {
		log.Fatal("Connect failed:", err)
	}
	defer client.Close()

	request := Request{
		Action: "echo",
		Data:   "Hello, World!",
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		resp, err := client.Send(request)
		if err != nil {
			log.Printf("Send failed: %v", err)
			continue
		}

		var response Response
		codec := v2.NewDefaultCodec()
		if err := codec.Decode(resp, &response); err != nil {
			log.Printf("Decode failed: %v", err)
			continue
		}

		if i%100 == 0 {
			log.Printf("Response: %+v", response)
		}
	}

	elapsed := time.Since(start)
	log.Printf("Processed 1000 requests in %v", elapsed)
}
