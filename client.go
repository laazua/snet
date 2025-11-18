package snet

import (
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Client TCP客户端
type Client struct {
	conn      *Conn
	addr      string
	seq       uint32
	mu        sync.Mutex
	connected bool
}

// NewClient 创建客户端
func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
		seq:  0,
	}
}

// Connect 连接服务器
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return ErrClientConnected
	}
	var err error
	var conn net.Conn
	if clientAuthConfig != nil {
		conn, err = tls.Dial("tcp", c.addr, clientAuthConfig)
	} else {
		conn, err = net.Dial("tcp", c.addr)
	}
	if err != nil {
		return err
	}

	c.conn = newConn(conn)
	c.connected = true

	return nil
}

// Send 发送数据
func (c *Client) Send(dataType PacketType, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return ErrClientNotConnected
	}

	seq := atomic.AddUint32(&c.seq, 1)
	packet := NewPacket(dataType, data, seq)
	return c.conn.SendPacket(packet)
}

// Receive 接收数据
func (c *Client) Receive() (*Packet, error) {
	if !c.connected {
		return nil, ErrClientNotConnected
	}

	return c.conn.ReceivePacket()
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected && c.conn != nil
}

// Reconnect 重新连接
func (c *Client) Reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.conn.Close()
		c.connected = false
	}

	var err error
	var conn net.Conn
	if clientAuthConfig != nil {
		conn, err = tls.Dial("tcp", c.addr, clientAuthConfig)
	} else {
		conn, err = net.Dial("tcp", c.addr)
	}
	if err != nil {
		return err
	}

	c.conn = newConn(conn)
	c.connected = true
	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.connected = false
		return c.conn.Close()
	}

	return nil
}

// SetTimeout 设置超时
func (c *Client) SetTimeout(readTimeout, writeTimeout time.Duration) {
	if c.conn != nil {
		c.conn.SetTimeout(readTimeout, writeTimeout)
	}
}
