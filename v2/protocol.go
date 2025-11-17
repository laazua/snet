package snet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

const (
	// 协议头
	headerSize     = 8
	magicNumber    = 0x1234ABCD
	maxMessageSize = 10 * 1024 * 1024 // 10MB
)

var (
	ErrMessageTooLarge = errors.New("message too large")
	ErrInvalidMagic    = errors.New("invalid magic number")
	ErrInvalidHeader   = errors.New("invalid header")
)

// PacketHeader 数据包头部
type PacketHeader struct {
	Magic     uint32
	Length    uint32
	MessageID uint32
}

// Protocol 协议处理器
type Protocol struct {
	bufferPool *BufferPool
	slicePool  *ByteSlicePool
}

// NewProtocol 创建协议处理器
func NewProtocol() *Protocol {
	return &Protocol{
		bufferPool: NewBufferPool(),
		slicePool:  NewByteSlicePool(headerSize),
	}
}

// Pack 打包消息
func (p *Protocol) Pack(msgID uint32, data interface{}) ([]byte, error) {
	// 序列化数据
	var bodyData []byte
	var err error

	switch v := data.(type) {
	case []byte:
		bodyData = v
	case string:
		bodyData = []byte(v)
	default:
		// 使用JSON序列化
		bodyData, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	// 检查消息大小
	if len(bodyData) > maxMessageSize {
		return nil, ErrMessageTooLarge
	}

	// 创建缓冲区
	totalSize := headerSize + len(bodyData)
	buffer := p.bufferPool.Get(totalSize)
	defer p.bufferPool.Put(buffer)

	// 写入头部
	header := PacketHeader{
		Magic:     magicNumber,
		Length:    uint32(len(bodyData)),
		MessageID: msgID,
	}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return nil, err
	}

	// 写入数据
	if _, err := buffer.Write(bodyData); err != nil {
		return nil, err
	}

	// 返回数据副本
	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())
	return result, nil
}

// Unpack 解包消息
func (p *Protocol) Unpack(reader io.Reader) (*Message, error) {
	// 读取头部
	headerBuf := p.slicePool.Get()
	defer p.slicePool.Put(headerBuf)

	if _, err := io.ReadFull(reader, headerBuf); err != nil {
		return nil, err
	}

	// 解析头部
	var header PacketHeader
	buffer := bytes.NewBuffer(headerBuf)
	if err := binary.Read(buffer, binary.BigEndian, &header); err != nil {
		return nil, err
	}

	// 验证魔术字
	if header.Magic != magicNumber {
		return nil, ErrInvalidMagic
	}

	// 验证消息长度
	if header.Length > maxMessageSize {
		return nil, ErrMessageTooLarge
	}

	// 读取消息体
	bodyBuf := make([]byte, header.Length)
	if _, err := io.ReadFull(reader, bodyBuf); err != nil {
		return nil, err
	}

	// 创建消息
	msg := &Message{
		ID:   header.MessageID,
		Data: bodyBuf, // 保持原始字节，在handler中按需解析
	}

	return msg, nil
}

// ParseData 解析数据到指定类型
func (p *Protocol) ParseData(data []byte, target interface{}) error {
	if data == nil {
		return nil
	}

	// 根据目标类型进行解析
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}

	targetType := targetValue.Elem().Type()

	switch targetType.Kind() {
	case reflect.String:
		targetValue.Elem().SetString(string(data))
		return nil
	case reflect.Slice:
		if targetType.Elem().Kind() == reflect.Uint8 {
			// []byte 类型
			targetValue.Elem().SetBytes(data)
			return nil
		}
	}

	// 默认使用JSON反序列化
	return json.Unmarshal(data, target)
}

// ConvertToBytes 将数据转换为字节切片
func (p *Protocol) ConvertToBytes(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}
