package snet

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

// Server TCP服务器
type Server struct {
	addr        string
	listener    net.Listener
	handler     Handler
	workerPool  *WorkerPool
	connManager *ConnManager
	mu          sync.RWMutex
	running     bool
}

// Handler 请求处理器接口
type Handler interface {
	Handle(conn *Conn, packet *Packet)
}

// HandlerFunc 处理器函数类型
type HandlerFunc func(conn *Conn, packet *Packet)

func (f HandlerFunc) Handle(conn *Conn, packet *Packet) {
	f(conn, packet)
}

// NewServer 创建服务器
func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
		// workerPool:  NewWorkerPool(workers, 1000),
		connManager: NewConnManager(),
	}
}

func (s *Server) SetHandler(handler Handler) *Server {
	s.handler = handler
	return s
}

func (s *Server) SetWorkerPool(workers, maxQueueSize int) *Server {
	s.workerPool = newWorkerPool(workers, maxQueueSize)
	return s
}

// Start 启动服务器
func (s *Server) Start() error {
	var err error
	var listener net.Listener
	if serverAuthConfig != nil {
		slog.Info("加载证书启动")
		listener, err = tls.Listen("tcp", s.addr, serverAuthConfig)
	} else {
		slog.Info("未加载证书启动")
		listener, err = net.Listen("tcp", s.addr)
	}
	if err != nil {
		return err
	}
	if s.handler == nil {
		return fmt.Errorf("handler is not set")
	}
	if s.workerPool == nil {
		return fmt.Errorf("worker pool is not set")
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	fmt.Printf("Server started on %s\n", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.RLock()
			running := s.running
			s.mu.RUnlock()

			if !running {
				break
			}
			continue
		}

		go s.handleConnection(conn)
	}

	return nil
}

// handleConnection 处理连接
func (s *Server) handleConnection(netConn net.Conn) {
	conn := newConn(netConn)
	s.connManager.Add(conn)
	defer s.connManager.Remove(conn)
	defer conn.Close()

	for {
		packet, err := conn.ReceivePacket()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Receive packet error: %v\n", err)
			}
			break
		}

		// 处理心跳包
		if packet.Header.Type == PacketTypeHeartbeat {
			ackPacket := NewPacket(PacketTypeAck, []byte("pong"), packet.Header.Seq)
			conn.SendPacket(ackPacket)
			continue
		}

		// 提交到协程池处理
		s.workerPool.Submit(func() {
			s.handler.Handle(conn, packet)
		})
	}
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.running = false
		s.listener.Close()
		s.workerPool.Close()
		s.connManager.CloseAll()
	}
}
