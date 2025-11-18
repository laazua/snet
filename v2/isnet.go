package v2

import (
	"context"
	"net"
	"time"
)

// Handler 业务处理接口
type Handler interface {
	Handle(ctx context.Context, req []byte) ([]byte, error)
}

// HandlerFunc 业务处理函数类型
type HandlerFunc func(ctx context.Context, req []byte) ([]byte, error)

func (f HandlerFunc) Handle(ctx context.Context, req []byte) ([]byte, error) {
	return f(ctx, req)
}

// Codec 编解码器接口
type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}

// Pool 协程池接口
type Pool interface {
	Submit(task func()) error
	Release()
}

// Conn 连接接口
type Conn interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

// Server 服务器接口
type Server interface {
	Start() error
	Stop() error
	RegisterHandler(handler Handler)
}

// Client 客户端接口
type Client interface {
	Connect(addr string) error
	Close() error
	Send(req interface{}) ([]byte, error)
	SendWithTimeout(req interface{}, timeout time.Duration) ([]byte, error)
}
