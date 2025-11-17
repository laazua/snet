package snet

import (
	"crypto/tls"
	"time"
)

// Config 服务器配置
type Config struct {
	Address         string        // 监听地址
	MaxConnections  int           // 最大连接数
	WorkerPoolSize  int           // 工作协程池大小
	MaxWorkerTasks  int           // 每个工作协程最大任务数
	ReadBufferSize  int           // 读缓冲区大小
	WriteBufferSize int           // 写缓冲区大小
	ReadTimeout     time.Duration // 读超时
	WriteTimeout    time.Duration // 写超时
	TLSConfig       *tls.Config   // TLS配置
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Address:         ":8080",
		MaxConnections:  10000,
		WorkerPoolSize:  1000,
		MaxWorkerTasks:  1000,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		TLSConfig:       nil,
	}
}

// RequestHandler 请求处理函数类型
type RequestHandler func(*Context) error

// Context 请求上下文
type Context struct {
	Conn    *Connection
	Request *Message
}

// Message 消息结构
type Message struct {
	ID   uint32
	Data interface{}
}

// Server 服务器接口
type Server interface {
	Start() error
	Stop() error
	RegisterHandler(msgID uint32, handler RequestHandler)
}

// Client 客户端接口
type Client interface {
	Connect() error
	Close() error
	Send(msgID uint32, data interface{}) error
	SendWithResponse(msgID uint32, data interface{}, timeout time.Duration) (*Message, error)
}

// NewServer 创建新的服务器
func NewServer(config *Config) Server {
	return newServer(config)
}

// NewClient 创建新的客户端
func NewClient(address string, tlsConfig *tls.Config) Client {
	return newClient(address, tlsConfig)
}
