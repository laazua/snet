package v2

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	Network      string        // 网络类型
	Timeout      time.Duration // 连接超时
	ReadTimeout  time.Duration // 读超时
	WriteTimeout time.Duration // 写超时
	TLSConfig    *tls.Config   // TLS配置
	EnableTLS    bool          // 是否启用TLS
}

// client TCP客户端
type client struct {
	config   *ClientConfig
	conn     Conn
	mu       sync.RWMutex
	codec    Codec
	isClosed bool
}

// NewClient 创建新的TCP客户端
func NewClient(config *ClientConfig, codec Codec) Client {
	if config == nil {
		config = &ClientConfig{
			Network:      "tcp",
			Timeout:      10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	return &client{
		config: config,
		codec:  codec,
	}
}

func (c *client) Connect(addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return errors.New("client already connected")
	}

	var netConn net.Conn
	var err error

	if c.config.EnableTLS {
		dialer := &tls.Dialer{
			Config: c.config.TLSConfig,
		}
		netConn, err = dialer.Dial(c.config.Network, addr)
	} else {
		netConn, err = net.DialTimeout(c.config.Network, addr, c.config.Timeout)
	}

	if err != nil {
		return err
	}

	c.conn = newConn(netConn)
	c.isClosed = false
	return nil
}

func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return nil
	}

	c.isClosed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *client) Send(req interface{}) ([]byte, error) {
	return c.SendWithTimeout(req, c.config.ReadTimeout)
}

func (c *client) SendWithTimeout(req interface{}, timeout time.Duration) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isClosed || c.conn == nil {
		return nil, errors.New("client is closed or not connected")
	}

	// 序列化请求
	data, err := c.codec.Encode(req)
	if err != nil {
		return nil, err
	}

	// 发送请求
	if err := c.conn.Write(data); err != nil {
		return nil, err
	}

	// 设置读取超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 读取响应
	resultChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		resp, err := c.conn.Read()
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- resp
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case resp := <-resultChan:
		return resp, nil
	}
}
