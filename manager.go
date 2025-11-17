package snet

import (
	"net"
	"sync"
)

// ConnManager 连接管理器
type ConnManager struct {
	conns map[net.Conn]*Conn
	mu    sync.RWMutex
}

// NewConnManager 创建连接管理器
func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make(map[net.Conn]*Conn),
	}
}

// Add 添加连接
func (cm *ConnManager) Add(conn *Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[conn.Conn] = conn
}

// Remove 移除连接
func (cm *ConnManager) Remove(conn *Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conns, conn.Conn)
}

// Get 获取连接
func (cm *ConnManager) Get(conn net.Conn) *Conn {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conns[conn]
}

// CloseAll 关闭所有连接
func (cm *ConnManager) CloseAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conn := range cm.conns {
		conn.Close()
	}
	cm.conns = make(map[net.Conn]*Conn)
}

// Count 连接数量
func (cm *ConnManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.conns)
}
