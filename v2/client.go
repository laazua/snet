package v2

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// Client 客户端
type Client struct {
	serverAddr  string
	conn        net.Conn
	isConnected bool
	serializer  Serializer
	tlsConfig   *tls.Config
}

// NewClient 创建客户端
func NewClient(serverAddr string, useTLS bool, certFile, keyFile string) *Client {
	client := &Client{
		serverAddr: serverAddr,
		serializer: NewJSONSerializer(),
	}

	if useTLS {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(err)
		}

		client.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   "localhost",
		}
	}

	return client
}

// Connect 连接服务器
func (c *Client) Connect() error {
	var err error
	if c.tlsConfig != nil {
		c.conn, err = tls.Dial("tcp", c.serverAddr, c.tlsConfig)
	} else {
		c.conn, err = net.Dial("tcp", c.serverAddr)
	}

	if err != nil {
		return err
	}

	c.isConnected = true
	return nil
}

// SendMsg 发送消息
func (c *Client) SendMsg(msgID uint32, data []byte) error {
	if !c.isConnected {
		return fmt.Errorf("client not connected")
	}

	msg := NewMessage(msgID, data)
	dataPack := NewDataPack()
	packedData, err := dataPack.Pack(msg)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(packedData)
	return err
}

// SendStruct 发送结构体
func (c *Client) SendStruct(msgID uint32, data interface{}) error {
	serializedData, err := c.serializer.Serialize(data)
	if err != nil {
		return err
	}
	return c.SendMsg(msgID, serializedData)
}

// ReadMsg 读取消息
func (c *Client) ReadMsg() (Message, error) {
	dataPack := NewDataPack()
	header := make([]byte, dataPack.GetHeadLen())

	if _, err := c.conn.Read(header); err != nil {
		return nil, err
	}

	msg, err := dataPack.Unpack(header)
	if err != nil {
		return nil, err
	}

	if msg.GetDataLen() > 0 {
		data := make([]byte, msg.GetDataLen())
		if _, err := c.conn.Read(data); err != nil {
			return nil, err
		}
		msg.SetData(data)
	}

	return msg, nil
}

// Close 关闭连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.isConnected = false
}

// SetDeadline 设置超时
func (c *Client) SetDeadline(duration time.Duration) error {
	return c.conn.SetDeadline(time.Now().Add(duration))
}
