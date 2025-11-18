package v2

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
)

// Server 服务器
type Server struct {
	Name        string
	IPVersion   string
	IP          string
	Port        int
	workerPool  WorkerPool
	router      *Router
	connMgr     *ConnectionManager
	serializer  Serializer
	tlsConfig   *tls.Config
	onConnStart func(conn Connection)
	onConnStop  func(conn Connection)
	sync.RWMutex
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	connections map[uint32]Connection
	sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[uint32]Connection),
	}
}

func (cm *ConnectionManager) Add(conn Connection) {
	cm.Lock()
	defer cm.Unlock()
	cm.connections[conn.GetConnID()] = conn
}

func (cm *ConnectionManager) Remove(connID uint32) {
	cm.Lock()
	defer cm.Unlock()
	delete(cm.connections, connID)
}

func (cm *ConnectionManager) Get(connID uint32) (Connection, bool) {
	cm.RLock()
	defer cm.RUnlock()
	conn, ok := cm.connections[connID]
	return conn, ok
}

func (cm *ConnectionManager) Len() int {
	cm.RLock()
	defer cm.RUnlock()
	return len(cm.connections)
}

func (cm *ConnectionManager) Clear() {
	cm.Lock()
	defer cm.Unlock()
	for connID, conn := range cm.connections {
		conn.Stop()
		delete(cm.connections, connID)
	}
}

// Server配置
type Config struct {
	Name           string
	Host           string
	Port           int
	WorkerNum      int
	MaxWorkerTask  int
	MaxConn        int
	UseTLS         bool
	CertFile       string
	KeyFile        string
	ClientAuthType tls.ClientAuthType
}

// NewServer 创建服务器
func NewServer(config *Config) *Server {
	s := &Server{
		Name:       config.Name,
		IPVersion:  "tcp4",
		IP:         config.Host,
		Port:       config.Port,
		router:     NewRouter(),
		connMgr:    NewConnectionManager(),
		serializer: NewJSONSerializer(),
	}

	// 初始化工作池
	if config.WorkerNum > 0 {
		s.workerPool = NewWorkerPool(config.WorkerNum, config.MaxWorkerTask)
	}

	// 配置TLS
	if config.UseTLS {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			panic(err)
		}

		s.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   config.ClientAuthType,
		}
	}

	return s
}

// Start 启动服务器
func (s *Server) Start() {
	fmt.Printf("[SNET] Server starting at %s:%d\n", s.IP, s.Port)

	// 启动工作池
	if s.workerPool != nil {
		s.workerPool.Start()
	}

	// 解析地址
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		panic(err)
	}

	// 监听端口
	var listener net.Listener
	if s.tlsConfig != nil {
		listener, err = tls.Listen(s.IPVersion, addr.String(), s.tlsConfig)
	} else {
		listener, err = net.ListenTCP(s.IPVersion, addr)
	}

	if err != nil {
		panic(err)
	}

	var connID uint32
	for {
		// 接受连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept err: %v\n", err)
			continue
		}

		// 限制最大连接数
		if s.connMgr.Len() >= 10000 { // 可根据配置调整
			conn.Close()
			continue
		}

		connID++
		dealConn := NewConnection(s, conn, connID)

		s.connMgr.Add(dealConn)
		go dealConn.Start()
	}
}

// Stop 停止服务器
func (s *Server) Stop() {
	fmt.Println("[SNET] Server stopping...")

	if s.workerPool != nil {
		s.workerPool.Stop()
	}

	s.connMgr.Clear()
}

// AddRouter 添加路由
func (s *Server) AddRouter(msgID uint32, handler RouterHandler) {
	s.router.AddRouter(msgID, handler)
}

// SetOnConnStart 设置连接开始回调
func (s *Server) SetOnConnStart(hookFunc func(Connection)) {
	s.onConnStart = hookFunc
}

// SetOnConnStop 设置连接结束回调
func (s *Server) SetOnConnStop(hookFunc func(Connection)) {
	s.onConnStop = hookFunc
}

// CallOnConnStart 调用连接开始回调
func (s *Server) CallOnConnStart(conn Connection) {
	if s.onConnStart != nil {
		s.onConnStart(conn)
	}
}

// CallOnConnStop 调用连接结束回调
func (s *Server) CallOnConnStop(conn Connection) {
	if s.onConnStop != nil {
		s.onConnStop(conn)
	}
	s.connMgr.Remove(conn.GetConnID())
}
