package main

import (
	"fmt"

	"github.com/laazua/snet"
)

func handleLogin(conn *snet.Conn, packet *snet.Packet) {
	// 处理登录逻辑
	fmt.Println("Handle login packet")
}

func handleChat(conn *snet.Conn, packet *snet.Packet) {
	// 处理聊天逻辑
	fmt.Println("Handle chat packet")
}

func handleFile(conn *snet.Conn, packet *snet.Packet) {
	// 处理文件传输逻辑
	fmt.Println("Handle file packet")
}

func handleDefault(conn *snet.Conn, packet *snet.Packet) {
	// 处理未知包类型
	fmt.Printf("Handle unknown packet type: %d\n", packet.Header.Type)
}
