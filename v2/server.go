package v2

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Network        string        // 网络类型 tcp/tcp4/tcp6
	Addr           string        // 监听地址
	MaxConnections int           // 最大连接数
	WorkerCount    int           // 工作协程数
	QueueSize      int           // 任务队列大小
	ReadTimeout    time.Duration // 读超时
	WriteTimeout   time.Duration // 写超时
	TLSConfig      *tls.Config   // TLS配置
	EnableTLS      bool          // 是否启用TLS
	CertFile       string        // 证书文件
	KeyFile        string        // 私钥文件
}

// server TCP服务器
type server struct {
	config      *ServerConfig
	listener    net.Listener
	handler     Handler
	pool        Pool
	wg          sync.WaitGroup
	mu          sync.RWMutex
	connections map[net.Conn]struct{}
	closed      int32
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewServer 创建新的TCP服务器
func NewServer(config *ServerConfig) Server {
	if config == nil {
		config = &ServerConfig{
			Network:        "tcp",
			MaxConnections: 10000,
			WorkerCount:    100,
			QueueSize:      1000,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &server{
		config:      config,
		handler:     nil,
		connections: make(map[net.Conn]struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *server) Start() error {
	var listener net.Listener
	var err error

	// 创建监听器
	if s.config.EnableTLS {
		if s.config.TLSConfig == nil {
			cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
			if err != nil {
				return err
			}
			s.config.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
			}
		}
		listener, err = tls.Listen(s.config.Network, s.config.Addr, s.config.TLSConfig)
	} else {
		listener, err = net.Listen(s.config.Network, s.config.Addr)
	}

	if err != nil {
		return err
	}

	s.listener = listener
	s.pool = NewWorkerPool(s.config.WorkerCount, s.config.QueueSize)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *server) Stop() error {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return nil
	}

	s.cancel()

	if s.listener != nil {
		s.listener.Close()
	}

	if s.pool != nil {
		s.pool.Release()
	}

	s.mu.Lock()
	for conn := range s.connections {
		conn.Close()
	}
	s.connections = make(map[net.Conn]struct{})
	s.mu.Unlock()

	s.wg.Wait()
	return nil
}

func (s *server) RegisterHandler(handler Handler) {
	s.handler = handler
}

func (s *server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if atomic.LoadInt32(&s.closed) == 1 {
				return
			}
			continue
		}

		s.mu.Lock()
		if len(s.connections) >= s.config.MaxConnections {
			conn.Close()
			s.mu.Unlock()
			continue
		}
		s.connections[conn] = struct{}{}
		s.mu.Unlock()

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *server) handleConnection(netConn net.Conn) {
	defer func() {
		netConn.Close()
		s.mu.Lock()
		delete(s.connections, netConn)
		s.mu.Unlock()
		s.wg.Done()
	}()

	conn := newConn(netConn)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		data, err := conn.Read()
		if err != nil {
			if err != io.EOF && !errors.Is(err, net.ErrClosed) {
				// 记录错误日志
			}
			return
		}

		if s.handler != nil {
			// 提交任务到协程池
			task := func() {
				ctx, cancel := context.WithTimeout(s.ctx, s.config.ReadTimeout)
				defer cancel()

				resp, err := s.handler.Handle(ctx, data)
				if err != nil {
					// 记录错误日志
					return
				}

				if resp != nil {
					if err := conn.Write(resp); err != nil {
						// 记录错误日志
					}
				}
			}

			if err := s.pool.Submit(task); err != nil {
				// 记录任务提交失败日志
			}
		}
	}
}
