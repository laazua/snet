package snet

import (
	"crypto/tls"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Connection TCP连接封装
type Connection struct {
	conn         net.Conn
	isClosed     int32
	readTimeout  time.Duration
	writeTimeout time.Duration
	protocol     *Protocol
	mu           sync.RWMutex
}

// NewConnection 创建新连接
func NewConnection(conn net.Conn, readTimeout, writeTimeout time.Duration) *Connection {
	return &Connection{
		conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		protocol:     NewProtocol(),
		isClosed:     0,
	}
}

// ReadMessage 读取消息
func (c *Connection) ReadMessage() (*Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if atomic.LoadInt32(&c.isClosed) == 1 {
		return nil, io.EOF
	}

	// 设置读超时
	if c.readTimeout > 0 {
		c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}

	return c.protocol.Unpack(c.conn)
}

// SendMessage 发送消息
func (c *Connection) SendMessage(msgID uint32, data interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if atomic.LoadInt32(&c.isClosed) == 1 {
		return io.EOF
	}

	// 打包消息
	packet, err := c.protocol.Pack(msgID, data)
	if err != nil {
		return err
	}

	// 设置写超时
	if c.writeTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	// 发送数据
	_, err = c.conn.Write(packet)
	return err
}

// Close 关闭连接
func (c *Connection) Close() error {
	if atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		return c.conn.Close()
	}
	return nil
}

// IsClosed 检查连接是否已关闭
func (c *Connection) IsClosed() bool {
	return atomic.LoadInt32(&c.isClosed) == 1
}

// RemoteAddr 返回远程地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr 返回本地地址
func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// SetDeadline 设置截止时间
func (c *Connection) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline 设置读截止时间
func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline 设置写截止时间
func (c *Connection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// wrapTLS 包装TLS连接
func wrapTLS(conn net.Conn, tlsConfig *tls.Config, isServer bool) (net.Conn, error) {
	if tlsConfig == nil {
		return conn, nil
	}

	if isServer {
		return tls.Server(conn, tlsConfig), nil
	}
	return tls.Client(conn, tlsConfig), nil
}
