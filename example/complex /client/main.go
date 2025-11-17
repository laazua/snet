// client_example.go
package main

import (
	"fmt"
	"log"
	"snet"
	"time"
)

type GameClient struct {
	client    snet.Client
	userID    int64
	token     string
	connected bool
}

func NewGameClient() *GameClient {
	return &GameClient{
		client: snet.NewClient("localhost:8888", nil),
	}
}

func (gc *GameClient) Run() error {
	// 连接服务器
	if err := gc.client.Connect(); err != nil {
		return err
	}
	gc.connected = true

	// 执行测试流程
	if err := gc.testLogin(); err != nil {
		return err
	}

	if err := gc.testChat(); err != nil {
		return err
	}

	if err := gc.testGameActions(); err != nil {
		return err
	}

	// 保持心跳
	gc.startHeartbeat()

	return nil
}

func (gc *GameClient) testLogin() error {
	fmt.Println("=== Testing Login ===")

	// 测试登录
	loginReq := UserLoginRequest{
		Username: "admin",
		Password: "123456",
		DeviceID: "device_001",
	}

	response, err := gc.client.SendWithResponse(MsgUserLogin, loginReq, 5*time.Second)
	if err != nil {
		return err
	}

	var loginResp UserLoginResponse
	if err := response.ParseData(&loginResp); err != nil {
		return err
	}

	if loginResp.Success {
		gc.userID = loginResp.UserID
		gc.token = loginResp.Token
		fmt.Printf("Login successful! UserID: %d, Token: %s\n", gc.userID, gc.token)
	} else {
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	// 获取用户信息
	userInfoReq := map[string]interface{}{
		"user_id": gc.userID,
	}

	infoResponse, err := gc.client.SendWithResponse(MsgUserInfo, userInfoReq, 5*time.Second)
	if err != nil {
		return err
	}

	var userInfo UserInfo
	if err := infoResponse.ParseData(&userInfo); err != nil {
		return err
	}

	fmt.Printf("User Info: %+v\n", userInfo)

	return nil
}

func (gc *GameClient) testChat() error {
	fmt.Println("\n=== Testing Chat ===")

	// 发送私聊消息
	chatMsg := ChatMessage{
		FromUserID: gc.userID,
		ToUserID:   10002, // 假设另一个用户
		Content:    "Hello from user 10001!",
		Type:       1,
		Timestamp:  time.Now().Unix(),
	}

	response, err := gc.client.SendWithResponse(MsgChatPrivate, chatMsg, 5*time.Second)
	if err != nil {
		return err
	}

	var chatResp ChatResponse
	if err := response.ParseData(&chatResp); err != nil {
		return err
	}

	if chatResp.Success {
		fmt.Printf("Chat message sent successfully! MessageID: %d\n", chatResp.MessageID)
	} else {
		fmt.Printf("Failed to send chat: %s\n", chatResp.ErrorMsg)
	}

	return nil
}

func (gc *GameClient) testGameActions() error {
	fmt.Println("\n=== Testing Game Actions ===")

	// 测试玩家移动
	for i := 0; i < 3; i++ {
		move := PlayerMove{
			UserID:    gc.userID,
			X:         float64(i * 10),
			Y:         float64(i * 5),
			Z:         0,
			Direction: float64(i * 90),
			Timestamp: time.Now().Unix(),
		}

		if err := gc.client.Send(MsgGameMove, move); err != nil {
			return err
		}

		fmt.Printf("Sent move: (%.1f, %.1f, %.1f) direction: %.1f\n",
			move.X, move.Y, move.Z, move.Direction)

		time.Sleep(500 * time.Millisecond)
	}

	// 测试游戏动作
	actions := []GameAction{
		{
			UserID:    gc.userID,
			Action:    "attack",
			Params:    map[string]interface{}{"damage": 150, "target": 10002},
			Timestamp: time.Now().Unix(),
		},
		{
			UserID:    gc.userID,
			Action:    "use_item",
			Params:    map[string]interface{}{"item_id": 1001, "count": 1},
			Timestamp: time.Now().Unix(),
		},
		{
			UserID:    gc.userID,
			Action:    "cast_skill",
			Params:    map[string]interface{}{"skill_id": 201, "target_id": 10002},
			Timestamp: time.Now().Unix(),
		},
	}

	for _, action := range actions {
		response, err := gc.client.SendWithResponse(MsgGameAction, action, 5*time.Second)
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err := response.ParseData(&result); err != nil {
			return err
		}

		fmt.Printf("Action %s result: %v\n", action.Action, result)
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (gc *GameClient) startHeartbeat() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !gc.connected {
				break
			}

			heartbeat := Heartbeat{
				ClientTime: time.Now().Unix(),
			}

			response, err := gc.client.SendWithResponse(MsgHeartbeat, heartbeat, 10*time.Second)
			if err != nil {
				log.Printf("Heartbeat failed: %v", err)
				continue
			}

			var systemInfo SystemInfo
			if err := response.ParseData(&systemInfo); err != nil {
				log.Printf("Parse heartbeat response failed: %v", err)
				continue
			}

			fmt.Printf("Heartbeat response - Online: %d, ServerTime: %d\n",
				systemInfo.OnlineUsers, systemInfo.ServerTime)
		}
	}()
}

func (gc *GameClient) Close() {
	if gc.connected {
		// 发送登出消息
		logoutReq := map[string]interface{}{
			"user_id": gc.userID,
			"token":   gc.token,
		}

		gc.client.Send(MsgUserLogout, logoutReq)
		gc.client.Close()
		gc.connected = false
	}
}

func main() {
	client := NewGameClient()
	defer client.Close()

	if err := client.Run(); err != nil {
		log.Fatal("Client run failed:", err)
	}

	// 保持客户端运行
	fmt.Println("\nClient running... Press Ctrl+C to exit.")
	select {}
}
