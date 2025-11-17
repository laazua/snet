package snet

import (
	"bytes"
	"sync"
)

const (
	smallBufferSize = 1024      // 1KB
	largeBufferSize = 32 * 1024 // 32KB
)

// BufferPool 缓冲区池
type BufferPool struct {
	smallPool *sync.Pool
	largePool *sync.Pool
}

// NewBufferPool 创建缓冲区池
func NewBufferPool() *BufferPool {
	return &BufferPool{
		smallPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, smallBufferSize))
			},
		},
		largePool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, largeBufferSize))
			},
		},
	}
}

// Get 获取缓冲区
func (p *BufferPool) Get(size int) *bytes.Buffer {
	if size <= smallBufferSize {
		buf := p.smallPool.Get().(*bytes.Buffer)
		buf.Reset()
		return buf
	} else if size <= largeBufferSize {
		buf := p.largePool.Get().(*bytes.Buffer)
		buf.Reset()
		return buf
	}

	// 大缓冲区直接分配
	return bytes.NewBuffer(make([]byte, 0, size))
}

// Put 归还缓冲区
func (p *BufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}

	cap := buf.Cap()
	if cap <= smallBufferSize {
		p.smallPool.Put(buf)
	} else if cap <= largeBufferSize {
		p.largePool.Put(buf)
	}
	// 大缓冲区让GC回收
}

// ByteSlicePool 字节切片池
type ByteSlicePool struct {
	pool *sync.Pool
}

// NewByteSlicePool 创建字节切片池
func NewByteSlicePool(size int) *ByteSlicePool {
	return &ByteSlicePool{
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

// Get 获取字节切片
func (p *ByteSlicePool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put 归还字节切片
func (p *ByteSlicePool) Put(b []byte) {
	if b != nil {
		p.pool.Put(b)
	}
}
