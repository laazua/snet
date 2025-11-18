package v2

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var (
	// ErrConnectionClosed 连接已关闭错误
	ErrConnectionClosed = errors.New("connection closed")
	// ErrWriteTimeout 写入超时错误
	ErrWriteTimeout = errors.New("write timeout")
)

// conn 连接封装
type conn struct {
	netConn    net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	mu         sync.RWMutex
	closed     bool
	writeChan  chan *writeRequest
	closeChan  chan struct{}
	bufferPool *sync.Pool
}

type writeRequest struct {
	data []byte
	err  chan error
}

// newConn 创建新连接
func newConn(netConn net.Conn) Conn {
	c := &conn{
		netConn:   netConn,
		reader:    bufio.NewReader(netConn),
		writer:    bufio.NewWriter(netConn),
		writeChan: make(chan *writeRequest, 1000),
		closeChan: make(chan struct{}),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 4096)
			},
		},
	}

	go c.writeLoop()
	return c
}

// newTLSConn 创建TLS连接
func newTLSConn(netConn net.Conn, config *tls.Config) (Conn, error) {
	tlsConn := tls.Client(netConn, config)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	return newConn(tlsConn), nil
}

func (c *conn) Read() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	packet, err := ReadPacket(c.reader)
	if err != nil {
		if err == io.EOF {
			c.close()
		}
		return nil, err
	}

	return packet.Data, nil
}

func (c *conn) Write(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	req := &writeRequest{
		data: data,
		err:  make(chan error, 1),
	}

	select {
	case c.writeChan <- req:
		return <-req.err
	case <-c.closeChan:
		return ErrConnectionClosed
	case <-time.After(5 * time.Second):
		return ErrWriteTimeout
	}
}

func (c *conn) writeLoop() {
	for {
		select {
		case req := <-c.writeChan:
			err := WritePacket(c.writer, req.data)
			if err == nil {
				err = c.writer.Flush()
			}
			req.err <- err
			if err != nil {
				c.close()
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

func (c *conn) Close() error {
	c.close()
	return c.netConn.Close()
}

func (c *conn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.closeChan)
	}
}

func (c *conn) LocalAddr() net.Addr {
	return c.netConn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.netConn.RemoteAddr()
}
