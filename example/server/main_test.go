package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/laazua/snet"
)

// go test -bench=. -benchtime=10s -benchmem
func BenchmarkTCPServer(b *testing.B) {

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client := snet.NewClient("localhost:8082")
			if err := client.Connect(); err != nil {
				log.Fatal(err)
			}
			defer client.Close()

			JsonReq(client)
		}
	})
}

func JsonReq(conn *snet.Client) {
	serializer := snet.NewDataSerializer()
	// 发送map数据
	mapData := map[string]any{
		"action": "login",
		"user":   "john",
		"time":   time.Now().Unix(),
	}

	mapBytes, err := serializer.SerializeMap(mapData)
	if err != nil {
		log.Fatal(err)
	}
	// 这里发送的数据包类型为JSON数据类型, 服务端必须添加了该类型的handler才能处理
	if err := conn.Send(snet.PacketTypeDataJson, mapBytes); err != nil {
		log.Fatal(err)
	}

	// 接收响应
	response, err := conn.Receive()
	if err != nil {
		log.Fatal(err)
	}

	responseMap, err := serializer.DeserializeMap(response.Data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received map response: %v\n", responseMap)
}

// 运行基准测试
// go test -bench=. -benchtime=10s -benchmem
