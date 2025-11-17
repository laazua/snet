// message_types.go
package complex

// 基础消息类型定义
const (
	// 用户相关消息 1000-1999
	MsgUserLogin    = 1001
	MsgUserRegister = 1002
	MsgUserLogout   = 1003
	MsgUserInfo     = 1004

	// 聊天相关消息 2000-2999
	MsgChatPrivate   = 2001
	MsgChatGroup     = 2002
	MsgChatBroadcast = 2003

	// 游戏相关消息 3000-3999
	MsgGameMove   = 3001
	MsgGameAction = 3002
	MsgGameState  = 3003

	// 系统消息 4000-4999
	MsgHeartbeat  = 4001
	MsgError      = 4002
	MsgSystemInfo = 4003
)

// 用户相关数据结构
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type UserLoginResponse struct {
	Success   bool   `json:"success"`
	UserID    int64  `json:"user_id"`
	Token     string `json:"token"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type UserInfo struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Level    int    `json:"level"`
	Exp      int64  `json:"exp"`
	Avatar   string `json:"avatar"`
}

// 聊天相关数据结构
type ChatMessage struct {
	FromUserID int64  `json:"from_user_id"`
	ToUserID   int64  `json:"to_user_id"` // 0表示群聊或广播
	Content    string `json:"content"`
	Type       int    `json:"type"` // 1:文本 2:图片 3:语音
	Timestamp  int64  `json:"timestamp"`
}

type ChatResponse struct {
	MessageID int64  `json:"message_id"`
	Success   bool   `json:"success"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// 游戏相关数据结构
type PlayerMove struct {
	UserID    int64   `json:"user_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Z         float64 `json:"z"`
	Direction float64 `json:"direction"`
	Timestamp int64   `json:"timestamp"`
}

type GameAction struct {
	UserID    int64                  `json:"user_id"`
	Action    string                 `json:"action"`
	TargetID  int64                  `json:"target_id,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// 系统消息结构
type Heartbeat struct {
	ClientTime int64 `json:"client_time"`
}

type SystemInfo struct {
	OnlineUsers int   `json:"online_users"`
	ServerTime  int64 `json:"server_time"`
	MemoryUsage int64 `json:"memory_usage"`
}
