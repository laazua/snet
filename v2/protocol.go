package v2

import (
	"encoding/binary"
	"errors"
	"io"
)

// Packet 数据包结构
type Packet struct {
	Length uint32 // 数据长度
	Data   []byte // 实际数据
}

// Message 消息结构体
type Message struct {
	ID   uint32 // 消息ID
	Data []byte // 消息体
}

const (
	// MaxPacketSize 最大数据包大小 (10MB)
	MaxPacketSize = 10 * 1024 * 1024
	// HeaderSize 包头大小
	HeaderSize = 4 // 4字节长度字段
)

var (
	// ErrPacketTooLarge 数据包过大错误
	ErrPacketTooLarge = errors.New("packet size exceeds maximum limit")
	// ErrInvalidPacket 无效数据包错误
	ErrInvalidPacket = errors.New("invalid packet")
)

// ReadPacket 从连接读取完整数据包
func ReadPacket(r io.Reader) (*Packet, error) {
	// 读取长度字段
	lenBuf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)
	if length > MaxPacketSize {
		return nil, ErrPacketTooLarge
	}

	if length == 0 {
		return nil, ErrInvalidPacket
	}

	// 读取数据
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return &Packet{
		Length: length,
		Data:   data,
	}, nil
}

// WritePacket 写入数据包到连接
func WritePacket(w io.Writer, data []byte) error {
	length := uint32(len(data))
	if length > MaxPacketSize {
		return ErrPacketTooLarge
	}

	// 写入长度字段
	lenBuf := make([]byte, HeaderSize)
	binary.BigEndian.PutUint32(lenBuf, length)

	if _, err := w.Write(lenBuf); err != nil {
		return err
	}

	// 写入数据
	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}
