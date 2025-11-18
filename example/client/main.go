package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/laazua/snet"
)

// 自定义数据结构
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var id int

func main() {
	// 设置客户端TLS认证
	snet.SetClientAuth(
		"../certs/ssl/ca.crt",
		"../certs/ssl/client.crt",
		"../certs/ssl/client.key",
	)

	client := snet.NewClient("localhost:8082")
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	flag.IntVar(&id, "id", 250, "client id")
	flag.Parse()

	go HeartBeatReq(client)
	processBusiness(client)
	time.Sleep(time.Second)
}

func processBusiness(client *snet.Client) {
	// 检查连接状态
	if !client.IsConnected() {
		fmt.Println("Connection is not available, reconnecting...")
		if err := client.Reconnect(); err != nil {
			log.Fatal("Reconnect failed:", err)
		}
	}

	// 执行业务
	JsonReq(client)
	StructReq(client)

	// 模拟长时间业务处理
	for i := 0; i < 10; i++ {
		if !client.IsConnected() {
			fmt.Println("Connection lost during business processing")
			break
		}

		// 业务逻辑
		time.Sleep(5 * time.Second)
		fmt.Printf("Business processing %d...\n", i+1)
	}
}

func HeartBeatReq(client *snet.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 检查连接状态
		if !client.IsConnected() {
			fmt.Println("Connection lost, attempting to reconnect...")
			if err := client.Reconnect(); err != nil {
				fmt.Println("Reconnect failed:", err)
				continue
			}
			fmt.Println("Reconnected successfully")
		}

		// 这里发送的数据包类型为心跳包类型，服务端(内置)会识别并处理
		if err := client.Send(snet.PacketTypeHeartbeat, []byte("Client ping ...")); err != nil {
			fmt.Println("Heartbeat send error:", err)
			continue
		}
		fmt.Println("Heartbeat sent")

		resp, err := client.Receive()
		if err != nil {
			fmt.Println("Heartbeat receive error:", err)
			// 标记连接为断开状态
			client.Close()
			continue
		}
		fmt.Println("Heartbeat response:", string(resp.Data))
	}
}

func StructReq(conn *snet.Client) {
	serializer := snet.NewDataSerializer()
	// 发送结构体数据
	user := User{
		ID:   id,
		Name: "Alice",
		Age:  25,
	}

	userBytes, err := serializer.SerializeStruct(user)
	if err != nil {
		log.Fatal(err)
	}
	// 这里发送的数据包类型为自定义结构体数据类型, 服务端必须添加了该类型的handler才能处理
	if err := conn.Send(snet.PacketTypeDataStruct, userBytes); err != nil {
		log.Fatal(err)
	}

	// 接收结构体响应
	userResponse, err := conn.Receive()
	if err != nil {
		log.Fatal(err)
	}

	var responseUser User
	if err := serializer.DeserializeStruct(userResponse.Data, &responseUser); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received struct response: %+v\n", responseUser)
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
