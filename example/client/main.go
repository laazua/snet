package main

import (
	"flag"
	"fmt"
	"log"

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
	// snet.SetClientAuth(
	// 	"../certs/ssl/ca.crt",
	// 	"../certs/ssl/client.crt",
	// 	"../certs/ssl/client.key",
	// )
	client := snet.NewClient("localhost:8082")
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	flag.IntVar(&id, "id", 250, "client id")
	flag.Parse()

	serializer := snet.NewDataSerializer()

	// // 发送map数据
	// mapData := map[string]any{
	// 	"action": "login",
	// 	"user":   "john",
	// 	"time":   time.Now().Unix(),
	// }

	// mapBytes, err := serializer.SerializeMap(mapData)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// if err := client.Send(snet.PacketTypeData, mapBytes); err != nil {
	// 	log.Fatal(err)
	// }

	// // 接收响应
	// response, err := client.Receive()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// responseMap, err := serializer.DeserializeMap(response.Data)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Received map response: %v\n", responseMap)

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

	if err := client.Send(snet.PacketTypeData, userBytes); err != nil {
		log.Fatal(err)
	}

	// 接收结构体响应
	userResponse, err := client.Receive()
	if err != nil {
		log.Fatal(err)
	}

	var responseUser User
	if err := serializer.DeserializeStruct(userResponse.Data, &responseUser); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received struct response: %+v\n", responseUser)
}
