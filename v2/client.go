package snet

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrClientNotConnected = errors.New("client is not connected")
	ErrRequestTimeout     = errors.New("request timeout")
)

// client 客户端实现
type client struct {
	address   string
	tlsConfig *tls.Config
	conn      *Connection
	mu        sync.RWMutex
	connected int32
	responses sync.Map
	sequence  uint32
}

// newClient 创建新客户端
func newClient(address string, tlsConfig *tls.Config) *client {
	return &client{
		address:   address,
		tlsConfig: tlsConfig,
		connected: 0,
		sequence:  0,
	}
}

// Connect 连接服务器
func (c *client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if atomic.LoadInt32(&c.connected) == 1 {
		return nil
	}

	// 建立连接
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	// 包装TLS
	tlsConn, err := wrapTLS(conn, c.tlsConfig, false)
	if err != nil {
		conn.Close()
		return err
	}

	c.conn = NewConnection(tlsConn, 30*time.Second, 30*time.Second)
	atomic.StoreInt32(&c.connected, 1)

	// 启动响应处理协程
	go c.handleResponses()

	return nil
}

// handleResponses 处理响应
func (c *client) handleResponses() {
	for atomic.LoadInt32(&c.connected) == 1 {
		msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// 查找等待的响应通道
		if ch, exists := c.responses.LoadAndDelete(msg.ID); exists {
			if responseChan, ok := ch.(chan *Message); ok {
				select {
				case responseChan <- msg:
				default:
				}
			}
		}
	}
}

// Send 发送消息（无响应）
func (c *client) Send(msgID uint32, data interface{}) error {
	if atomic.LoadInt32(&c.connected) == 0 {
		return ErrClientNotConnected
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn.SendMessage(msgID, data)
}

// SendWithResponse 发送消息并等待响应
func (c *client) SendWithResponse(msgID uint32, data interface{}, timeout time.Duration) (*Message, error) {
	if atomic.LoadInt32(&c.connected) == 0 {
		return nil, ErrClientNotConnected
	}

	// 生成序列号
	sequence := atomic.AddUint32(&c.sequence, 1)
	responseID := msgID + sequence

	// 创建响应通道
	responseChan := make(chan *Message, 1)
	c.responses.Store(responseID, responseChan)
	defer c.responses.Delete(responseID)

	// 发送消息
	c.mu.RLock()
	err := c.conn.SendMessage(responseID, data)
	c.mu.RUnlock()

	if err != nil {
		return nil, err
	}

	// 等待响应
	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(timeout):
		return nil, ErrRequestTimeout
	}
}

// Close 关闭客户端
func (c *client) Close() error {
	if !atomic.CompareAndSwapInt32(&c.connected, 1, 0) {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}
