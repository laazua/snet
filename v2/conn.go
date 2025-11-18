package v2

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
)

// ConnectionImpl 连接实现
type ConnectionImpl struct {
	Server      *Server
	Conn        net.Conn
	ConnID      uint32
	isClosed    bool
	ctx         context.Context
	cancel      context.CancelFunc
	msgChan     chan []byte
	msgBuffChan chan []byte
	sync.RWMutex
}

// NewConnection 创建连接
func NewConnection(server *Server, conn net.Conn, connID uint32) *ConnectionImpl {
	c := &ConnectionImpl{
		Server:      server,
		Conn:        conn,
		ConnID:      connID,
		isClosed:    false,
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, 1024),
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func (c *ConnectionImpl) Start() {
	go c.startReader()
	go c.startWriter()

	c.Server.CallOnConnStart(c)
}

func (c *ConnectionImpl) Stop() {
	c.Lock()
	defer c.Unlock()

	if c.isClosed {
		return
	}

	c.Server.CallOnConnStop(c)
	c.cancel()
	close(c.msgChan)
	close(c.msgBuffChan)
	c.Conn.Close()
	c.isClosed = true
}

func (c *ConnectionImpl) Context() context.Context {
	return c.ctx
}

func (c *ConnectionImpl) GetConn() net.Conn {
	return c.Conn
}

func (c *ConnectionImpl) GetConnID() uint32 {
	return c.ConnID
}

func (c *ConnectionImpl) SendMsg(msgID uint32, data []byte) error {
	c.RLock()
	defer c.RUnlock()

	if c.isClosed {
		return errors.New("connection closed")
	}

	msg := NewMessage(msgID, data)
	dataPack := NewDataPack()
	packedData, err := dataPack.Pack(msg)
	if err != nil {
		return err
	}

	c.msgChan <- packedData
	return nil
}

func (c *ConnectionImpl) SendMsgWithStruct(msgID uint32, data interface{}) error {
	serializedData, err := c.Server.serializer.Serialize(data)
	if err != nil {
		return err
	}
	return c.SendMsg(msgID, serializedData)
}

func (c *ConnectionImpl) startReader() {
	defer c.Stop()

	dataPack := NewDataPack()
	header := make([]byte, dataPack.GetHeadLen())

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 读取头部
			if _, err := io.ReadFull(c.Conn, header); err != nil {
				return
			}

			// 解析头部
			msg, err := dataPack.Unpack(header)
			if err != nil {
				return
			}

			// 读取数据体
			data := make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.Conn, data); err != nil {
				return
			}
			msg.SetData(data)

			// 提交到工作池处理
			req := &RequestImpl{
				conn: c,
				msg:  msg,
			}

			if c.Server.workerPool != nil {
				c.Server.workerPool.Submit(func() {
					c.Server.router.Handle(req)
				})
			} else {
				go c.Server.router.Handle(req)
			}
		}
	}
}

func (c *ConnectionImpl) startWriter() {
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					return
				}
			} else {
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}
