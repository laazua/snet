package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/laazua/snet"
)

func main() {
	// 设置服务器端TLS认证
	snet.SetServerAuth(
		"../certs/ssl/ca.crt",
		"../certs/ssl/server.crt",
		"../certs/ssl/server.key",
	)

	server := snet.NewServer(":8082").SetWorkerPool(4, 1000)

	// 注册数据包处理函数(这里注册的数据包类型需要和客户端发送的数据包类型一致)
	// server.AddHandlerFunc(snet.PacketTypeAuth, handleLogin)
	// server.AddHandlerFunc(snet.PacketTypeChat, handleChat)
	// server.AddHandlerFunc(snet.PacketTypeFile, handleFile)

	server.AddHandlerFunc(snet.PacketTypeDataJson, JsonHandler)
	server.AddHandlerFunc(snet.PacketTypeDataStruct, StructHandler)

	// 注册信号处理函数
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			fmt.Println("Server error:", err)
			return
		}
	}()
	// 监听信号以优雅关闭服务器
	<-quit
	log.Println("Shutting down server...")
	server.Stop()
}

// json数据结构
func JsonHandler(conn *snet.Conn, packet *snet.Packet) {
	serializer := snet.NewDataSerializer()

	// 处理map数据
	data, err := serializer.DeserializeMap(packet.Data)
	if err != nil {
		return
	}
	// 收到数据进行业务逻辑处理
	fmt.Printf("Received map data: %v\n", data)
	// 返回处理后的结果
	resp := map[string]any{
		"status":  "success",
		"message": "Map data received",
		"data":    nil,
	}
	responseData, _ := serializer.SerializeMap(resp)
	responsePacket := snet.NewPacket(snet.PacketTypeDataJson, responseData, packet.Header.Seq)
	conn.SendPacket(responsePacket)
}

// 自定义数据结构
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func StructHandler(conn *snet.Conn, packet *snet.Packet) {
	serializer := snet.NewDataSerializer()
	var user User
	err := serializer.DeserializeStruct(packet.Data, &user)
	if err != nil {
		return
	}
	fmt.Printf("Received struct data: %+v\n", user)

	userResponse := User{
		ID:   user.ID,
		Name: "Response: " + user.Name,
		Age:  user.Age + 1,
	}
	responseData, _ := serializer.SerializeStruct(userResponse)
	responsePacket := snet.NewPacket(snet.PacketTypeDataStruct, responseData, packet.Header.Seq)
	conn.SendPacket(responsePacket)
}
