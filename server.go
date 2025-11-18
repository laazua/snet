package snet

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Server TCP服务器
type Server struct {
	addr           string
	listener       net.Listener
	handlers       map[PacketType]Handler // 基于包类型的handler映射
	defaultHandler Handler                // 默认handler
	workerPool     *WorkerPool
	connManager    *ConnManager
	mu             sync.RWMutex
	running        bool
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
		addr:        addr,
		handlers:    make(map[PacketType]Handler),
		workerPool:  newWorkerPool(100, 1000),
		connManager: NewConnManager(),
	}
}

// SetHandler 设置默认handler（兼容旧版本）
func (s *Server) SetHandler(handler Handler) *Server {
	s.defaultHandler = handler
	return s
}

// AddHandler 添加基于包类型的handler
func (s *Server) AddHandler(packetType PacketType, handler Handler) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[packetType] = handler
	return s
}

// AddHandlerFunc 添加基于包类型的handler函数
func (s *Server) AddHandlerFunc(packetType PacketType, handlerFunc func(conn *Conn, packet *Packet)) *Server {
	return s.AddHandler(packetType, HandlerFunc(handlerFunc))
}

// SetWorkerPool 设置工作池
func (s *Server) SetWorkerPool(workers, maxQueueSize int) *Server {
	s.workerPool = newWorkerPool(workers, maxQueueSize)
	return s
}

// getHandler 获取对应的handler
func (s *Server) getHandler(packetType PacketType) Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if handler, exists := s.handlers[packetType]; exists {
		return handler
	}
	return s.defaultHandler
}

// Start 启动服务器
func (s *Server) Start() error {
	var err error
	var listener net.Listener
	if serverAuthConfig != nil {
		log.Println("加载证书启动")
		listener, err = tls.Listen("tcp", s.addr, serverAuthConfig)
	} else {
		log.Println("未加载证书启动")
		listener, err = net.Listen("tcp", s.addr)
	}
	if err != nil {
		return err
	}

	// 检查是否有handler设置
	if len(s.handlers) == 0 && s.defaultHandler == nil {
		return ErrServerHandlerNotSet
	}
	if s.workerPool == nil {
		return ErrServerWorkerPoolNotSet
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	log.Printf("Server started on %s\n", s.addr)

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
	// 设置连接超时
	netConn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn := newConn(netConn)
	s.connManager.Add(conn)
	defer s.connManager.Remove(conn)
	defer conn.Close()

	for {
		packet, err := conn.ReceivePacket()
		if err != nil {
			if err != io.EOF {
				// fmt.Printf("Receive packet error: %v\n", err)
				// 检查是否是超时错误
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					log.Printf("Connection timeout: %v\n", err)
				} else {
					log.Printf("Receive packet error: %v\n", err)
				}
			}
			break
		}
		// 每次成功接收数据后重置超时时间
		netConn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// 处理心跳包
		if packet.Header.Type == PacketTypeHeartbeat {
			log.Println(string(packet.Data))

			ackPacket := NewPacket(PacketTypeAck, []byte("Server Pong ..."), packet.Header.Seq)
			conn.SendPacket(ackPacket)
			continue
		}

		// 获取对应的handler
		handler := s.getHandler(packet.Header.Type)
		if handler == nil {
			log.Printf("No handler found for packet type: %d\n", packet.Header.Type)
			continue
		}

		// 提交到协程池处理
		s.workerPool.Submit(func() {
			handler.Handle(conn, packet)
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
