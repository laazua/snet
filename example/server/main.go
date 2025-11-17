package main

import (
	"fmt"

	"github.com/laazua/snet"
)

func main() {
	// 设置服务器端TLS认证
	// snet.SetServerAuth(
	// 	"../certs/ssl/ca.crt",
	// 	"../certs/ssl/server.crt",
	// 	"../certs/ssl/server.key",
	// )
	handler := snet.HandlerFunc(StructHandler)
	server := snet.NewServer(":8082").SetHandler(handler).SetWorkerPool(4, 1000)
	// 注册信号处理函数

	// 启动服务器
	if err := server.Start(); err != nil {
		fmt.Println("Server error:", err)
		return
	}
	// 监听信号以优雅关闭服务器
	// server.Stop()
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
	responsePacket := snet.NewPacket(snet.PacketTypeData, responseData, packet.Header.Seq)
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
	responsePacket := snet.NewPacket(snet.PacketTypeData, responseData, packet.Header.Seq)
	conn.SendPacket(responsePacket)
}
