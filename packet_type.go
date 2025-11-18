package snet

// PacketType 数据包类型
type PacketType uint16

// 系统相关包类型 (0-99)
const (
	PacketTypeHeartbeat  PacketType = 1 // 心跳包
	PacketTypeHandshake  PacketType = 2 // 握手
	PacketTypeDisconnect PacketType = 3 // 断开连接
	PacketTypeAck        PacketType = 4 // 确认包
	PacketTypeDataJson   PacketType = 5 // json数据包
	PacketTypeDataStruct PacketType = 6 // 结构化数据包
	PacketTypeError      PacketType = 7 // 错误包
)

// 认证相关包类型 (100-199)
const (
	PacketTypeLogin  PacketType = 101 // 登录
	PacketTypeLogout PacketType = 102 // 登出
	PacketTypeAuth   PacketType = 103 // 认证
)

// 消息相关包类型 (200-299)
const (
	PacketTypeChat         PacketType = 201 // 聊天消息
	PacketTypeGroupChat    PacketType = 202 // 群聊消息
	PacketTypeBroadcast    PacketType = 203 // 广播消息
	PacketTypeNotification PacketType = 204 // 通知
)

// 文件相关包类型 (300-399)
const (
	PacketTypeFile      PacketType = 301 // 文件传输
	PacketTypeImage     PacketType = 302 // 图片
	PacketTypeVideo     PacketType = 303 // 视频
	PacketTypeAudio     PacketType = 304 // 音频
	PacketTypeFileStart PacketType = 305 // 文件开始传输
	PacketTypeFileData  PacketType = 306 // 文件数据
	PacketTypeFileEnd   PacketType = 307 // 文件传输结束
)

// 命令相关包类型 (400-499)
const (
	PacketTypeCommand PacketType = 401 // 命令
	PacketTypeQuery   PacketType = 402 // 查询
	PacketTypeUpdate  PacketType = 403 // 更新
	PacketTypeConfig  PacketType = 404 // 配置
)

// 业务相关包类型 (500-999)
const (
	PacketTypeLocation PacketType = 501 // 位置信息
	PacketTypeStatus   PacketType = 502 // 状态更新
	PacketTypePayment  PacketType = 503 // 支付
	// 可以根据具体业务继续扩展...
)
