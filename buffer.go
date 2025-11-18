package snet

import (
	"bytes"
	"sync"
)

// ==================== 字节缓冲区池 ====================
var bytesBufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func GetBytesBuffer() *bytes.Buffer {
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBytesBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if buf.Cap() > 1024*1024*2 {
		return
	}
	buf.Reset()
	bytesBufferPool.Put(buf)
}

// ==================== 字节切片池（指针版本） ====================
var byteSlicePool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

func GetByteSlice() []byte {
	ptr := byteSlicePool.Get().(*[]byte)
	*ptr = (*ptr)[:0] // 重置长度
	return *ptr
}

func PutByteSlice(b []byte) {
	if b == nil || cap(b) > 1024*1024 {
		return
	}
	b = b[:0] // 重置长度
	byteSlicePool.Put(&b)
}
