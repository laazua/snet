package snet

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrServerStopped = errors.New("server is stopped")
)

// server 服务器实现
type server struct {
	config      *Config
	listener    net.Listener
	workerPool  *WorkerPool
	handlers    map[uint32]RequestHandler
	connections sync.Map
	wg          sync.WaitGroup
	mu          sync.RWMutex
	stopped     int32
}

// newServer 创建新服务器
func newServer(config *Config) *server {
	return &server{
		config:   config,
		handlers: make(map[uint32]RequestHandler),
		stopped:  0,
	}
}

// Start 启动服务器
func (s *server) Start() error {
	// 创建监听器
	var listener net.Listener
	var err error

	if s.config.TLSConfig != nil {
		listener, err = tls.Listen("tcp", s.config.Address, s.config.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.config.Address)
	}

	if err != nil {
		return err
	}

	s.listener = listener

	// 初始化工作池
	s.workerPool = NewWorkerPool(s.config.WorkerPoolSize, s.config.MaxWorkerTasks)

	fmt.Printf("Server started on %s\n", s.config.Address)

	// 开始接受连接
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// acceptLoop 接受连接循环
func (s *server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if atomic.LoadInt32(&s.stopped) == 1 {
				return
			}
			continue
		}

		// 检查连接数限制
		if s.getConnectionCount() >= s.config.MaxConnections {
			conn.Close()
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (s *server) handleConnection(rawConn net.Conn) {
	defer s.wg.Done()

	// 包装TLS
	conn, err := wrapTLS(rawConn, s.config.TLSConfig, true)
	if err != nil {
		rawConn.Close()
		return
	}

	// 创建连接对象
	connection := NewConnection(conn, s.config.ReadTimeout, s.config.WriteTimeout)

	// 存储连接
	s.connections.Store(connection, struct{}{})
	defer s.connections.Delete(connection)

	// 处理消息
	for {
		if atomic.LoadInt32(&s.stopped) == 1 {
			break
		}

		msg, err := connection.ReadMessage()
		if err != nil {
			break
		}

		// 提交任务到工作池
		task := &serverTask{
			server:  s,
			conn:    connection,
			message: msg,
		}

		if err := s.workerPool.Submit(task); err != nil {
			// 工作池已满，直接处理
			task.Execute()
		}
	}

	connection.Close()
}

// serverTask 服务器任务
type serverTask struct {
	server  *server
	conn    *Connection
	message *Message
}

// Execute 执行任务
func (t *serverTask) Execute() {
	handler, exists := t.server.handlers[t.message.ID]
	if !exists {
		return
	}

	context := &Context{
		Conn:    t.conn,
		Request: t.message,
	}

	// 执行处理函数
	if err := handler(context); err != nil {
		// 处理错误
		fmt.Printf("Handler error: %v\n", err)
	}
}

// getConnectionCount 获取当前连接数
func (s *server) getConnectionCount() int {
	count := 0
	s.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Stop 停止服务器
func (s *server) Stop() error {
	if !atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		return nil
	}

	// 关闭监听器
	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有连接
	s.connections.Range(func(key, value interface{}) bool {
		if conn, ok := key.(*Connection); ok {
			conn.Close()
		}
		return true
	})

	// 关闭工作池
	if s.workerPool != nil {
		s.workerPool.Close()
	}

	// 等待所有连接关闭
	s.wg.Wait()

	return nil
}

// RegisterHandler 注册消息处理器
func (s *server) RegisterHandler(msgID uint32, handler RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[msgID] = handler
}
