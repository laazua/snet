package v2

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	// DefaultHeaderLen 默认头部长度 (ID 4字节 + DataLen 4字节)
	DefaultHeaderLen = 8
	// MaxPacketSize 最大数据包大小 10MB
	MaxPacketSize = 10 * 1024 * 1024
)

// DataPackImpl 数据包实现
type DataPackImpl struct{}

// NewDataPack 创建数据包实例
func NewDataPack() DataPack {
	return &DataPackImpl{}
}

func (dp *DataPackImpl) GetHeadLen() uint32 {
	return DefaultHeaderLen
}

func (dp *DataPackImpl) Pack(msg Message) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})

	// 写入DataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	// 写入MsgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	// 写入数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

func (dp *DataPackImpl) Unpack(binaryData []byte) (Message, error) {
	dataBuff := bytes.NewReader(binaryData)

	msg := &BaseMessage{}

	// 读取DataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读取MsgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	// 检查数据包是否超过最大限制
	if msg.DataLen > MaxPacketSize {
		return nil, errors.New("too large msg data received")
	}

	return msg, nil
}
