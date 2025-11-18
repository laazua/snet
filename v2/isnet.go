package v2

import (
	"context"
	"net"
)

// Message 消息接口
type Message interface {
	GetMsgID() uint32
	GetData() []byte
	SetData([]byte)
	GetDataLen() uint32
}

// Request 请求接口
type Request interface {
	GetConnection() Connection
	GetMessage() Message
}

// Connection 连接接口
type Connection interface {
	Start()
	Stop()
	Context() context.Context
	GetConn() net.Conn
	GetConnID() uint32
	SendMsg(msgID uint32, data []byte) error
	SendMsgWithStruct(msgID uint32, data interface{}) error
}

// RouterHandler 路由处理器接口
type RouterHandler interface {
	PreHandle(request Request)
	Handle(request Request)
	PostHandle(request Request)
}

// DataPack 数据包接口
type DataPack interface {
	GetHeadLen() uint32
	Pack(msg Message) ([]byte, error)
	Unpack([]byte) (Message, error)
}

// WorkerPool 工作池接口
type WorkerPool interface {
	Start()
	Stop()
	Submit(task func())
}
