package snet

import (
	"net"
	"sync"
	"time"
)

// Conn 连接封装
type Conn struct {
	net.Conn
	encoder      encoder
	decoder      decoder
	readTimeout  time.Duration
	writeTimeout time.Duration
	mu           sync.Mutex
}

// NewConn 创建连接
func newConn(conn net.Conn) *Conn {
	return &Conn{
		Conn:         conn,
		encoder:      &defaultEncoder{},
		decoder:      &defaultDecoder{},
		readTimeout:  30 * time.Second,
		writeTimeout: 30 * time.Second,
	}
}

// SetTimeout 设置超时时间
func (c *Conn) SetTimeout(readTimeout, writeTimeout time.Duration) {
	c.readTimeout = readTimeout
	c.writeTimeout = writeTimeout
}

// SendPacket 发送数据包
func (c *Conn) SendPacket(packet *Packet) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := c.encoder.encode(packet)
	if err != nil {
		return err
	}

	if c.writeTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	_, err = c.Conn.Write(data)
	return err
}

// ReceivePacket 接收数据包
func (c *Conn) ReceivePacket() (*Packet, error) {
	if c.readTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}

	return c.decoder.decode(c.Conn)
}

// Close 关闭连接
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.Close()
}
