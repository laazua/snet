// server_example.go
package main

import (
	"fmt"
	"log"
	"snet"
	"sync"
	"time"
)

type GameServer struct {
	server    snet.Server
	userConns sync.Map // userID -> *snet.Connection
	roomConns sync.Map // roomID -> map[userID]*snet.Connection
}

func NewGameServer() *GameServer {
	config := snet.DefaultConfig()
	config.Address = ":8888"
	config.WorkerPoolSize = 2000
	config.MaxConnections = 50000

	return &GameServer{
		server: snet.NewServer(config),
	}
}

func (gs *GameServer) Start() error {
	// 注册所有消息处理器
	gs.registerHandlers()

	// 启动服务器
	return gs.server.Start()
}

func (gs *GameServer) registerHandlers() {
	// 用户相关处理器
	gs.server.RegisterHandler(MsgUserLogin, gs.handleUserLogin)
	gs.server.RegisterHandler(MsgUserLogout, gs.handleUserLogout)
	gs.server.RegisterHandler(MsgUserInfo, gs.handleUserInfo)

	// 聊天相关处理器
	gs.server.RegisterHandler(MsgChatPrivate, gs.handlePrivateChat)
	gs.server.RegisterHandler(MsgChatGroup, gs.handleGroupChat)

	// 游戏相关处理器
	gs.server.RegisterHandler(MsgGameMove, gs.handlePlayerMove)
	gs.server.RegisterHandler(MsgGameAction, gs.handleGameAction)

	// 系统消息处理器
	gs.server.RegisterHandler(MsgHeartbeat, gs.handleHeartbeat)
}

// 用户登录处理
func (gs *GameServer) handleUserLogin(ctx *snet.Context) error {
	var req UserLoginRequest
	if err := ctx.Request.ParseData(&req); err != nil {
		return gs.sendError(ctx, "Invalid request format")
	}

	// 模拟用户验证
	userID, success := gs.authenticateUser(req.Username, req.Password)

	response := UserLoginResponse{
		Success:   success,
		UserID:    userID,
		Token:     "mock_token_" + req.Username,
		Message:   "Login processed",
		Timestamp: time.Now().Unix(),
	}

	if success {
		// 保存用户连接
		gs.userConns.Store(userID, ctx.Conn)
		fmt.Printf("User %s (ID: %d) logged in\n", req.Username, userID)
	}

	return ctx.Conn.SendMessage(ctx.Request.ID, response)
}

// 用户登出处理
func (gs *GameServer) handleUserLogout(ctx *snet.Context) error {
	var req map[string]interface{}
	if err := ctx.Request.ParseData(&req); err != nil {
		return err
	}

	userID := int64(req["user_id"].(float64))
	gs.userConns.Delete(userID)

	fmt.Printf("User ID: %d logged out\n", userID)
	return ctx.Conn.SendMessage(ctx.Request.ID, map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

// 私聊处理
func (gs *GameServer) handlePrivateChat(ctx *snet.Context) error {
	var chatMsg ChatMessage
	if err := ctx.Request.ParseData(&chatMsg); err != nil {
		return gs.sendError(ctx, "Invalid chat message format")
	}

	// 查找接收者连接
	if conn, exists := gs.userConns.Load(chatMsg.ToUserID); exists {
		// 转发消息给接收者
		if err := conn.(*snet.Connection).SendMessage(MsgChatPrivate, chatMsg); err != nil {
			return gs.sendError(ctx, "Failed to deliver message")
		}

		// 发送回执给发送者
		return ctx.Conn.SendMessage(ctx.Request.ID, ChatResponse{
			MessageID: time.Now().UnixNano(),
			Success:   true,
			Timestamp: time.Now().Unix(),
		})
	}

	return gs.sendError(ctx, "Recipient not online")
}

// 玩家移动处理
func (gs *GameServer) handlePlayerMove(ctx *snet.Context) error {
	var move PlayerMove
	if err := ctx.Request.ParseData(&move); err != nil {
		return gs.sendError(ctx, "Invalid move data")
	}

	// 处理移动逻辑
	fmt.Printf("Player %d moved to (%.2f, %.2f, %.2f)\n",
		move.UserID, move.X, move.Y, move.Z)

	// 广播给其他玩家（在实际游戏中需要根据视野范围广播）
	gs.broadcastToNearbyPlayers(move)

	return nil
}

// 游戏动作处理
func (gs *GameServer) handleGameAction(ctx *snet.Context) error {
	var action GameAction
	if err := ctx.Request.ParseData(&action); err != nil {
		return gs.sendError(ctx, "Invalid action data")
	}

	fmt.Printf("Player %d performed action: %s\n", action.UserID, action.Action)

	// 处理游戏逻辑
	switch action.Action {
	case "attack":
		gs.handleAttackAction(action)
	case "use_item":
		gs.handleUseItemAction(action)
	case "cast_skill":
		gs.handleCastSkillAction(action)
	}

	return ctx.Conn.SendMessage(ctx.Request.ID, map[string]interface{}{
		"success": true,
		"action":  action.Action,
		"result":  "completed",
	})
}

// 心跳处理
func (gs *GameServer) handleHeartbeat(ctx *snet.Context) error {
	var heartbeat Heartbeat
	if err := ctx.Request.ParseData(&heartbeat); err != nil {
		return err
	}

	// 返回服务器状态信息
	onlineUsers := 0
	gs.userConns.Range(func(_, _ interface{}) bool {
		onlineUsers++
		return true
	})

	return ctx.Conn.SendMessage(ctx.Request.ID, SystemInfo{
		OnlineUsers: onlineUsers,
		ServerTime:  time.Now().Unix(),
		MemoryUsage: 0, // 实际应该获取真实内存使用
	})
}

// 工具方法
func (gs *GameServer) authenticateUser(username, password string) (int64, bool) {
	// 模拟用户验证
	if username == "admin" && password == "123456" {
		return 10001, true
	}
	if username == "user1" && password == "111111" {
		return 10002, true
	}
	return 0, false
}

func (gs *GameServer) sendError(ctx *snet.Context, errorMsg string) error {
	return ctx.Conn.SendMessage(MsgError, map[string]interface{}{
		"error_code": 1,
		"error_msg":  errorMsg,
		"timestamp":  time.Now().Unix(),
	})
}

func (gs *GameServer) broadcastToNearbyPlayers(move PlayerMove) {
	// 在实际游戏中，这里应该只广播给附近的玩家
	gs.userConns.Range(func(key, value interface{}) bool {
		userID := key.(int64)
		conn := value.(*snet.Connection)

		if userID != move.UserID {
			// 发送移动信息给其他玩家
			conn.SendMessage(MsgGameMove, move)
		}
		return true
	})
}

func (gs *GameServer) handleAttackAction(action GameAction) {
	// 处理攻击逻辑
	damage := 100
	if val, exists := action.Params["damage"]; exists {
		damage = int(val.(float64))
	}

	fmt.Printf("Attack damage: %d\n", damage)
}

func (gs *GameServer) handleUseItemAction(action GameAction) {
	itemID := int64(action.Params["item_id"].(float64))
	fmt.Printf("Use item: %d\n", itemID)
}

func (gs *GameServer) handleCastSkillAction(action GameAction) {
	skillID := int(action.Params["skill_id"].(float64))
	targetID := int64(action.Params["target_id"].(float64))
	fmt.Printf("Cast skill %d on target %d\n", skillID, targetID)
}

func main() {
	server := NewGameServer()

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// 保持服务器运行
	select {}
}
