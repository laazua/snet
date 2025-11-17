package snet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Encoder 编码器接口
type encoder interface {
	encode(packet *Packet) ([]byte, error)
}

// Decoder 解码器接口
type decoder interface {
	decode(reader io.Reader) (*Packet, error)
}

// DefaultEncoder 默认编码器
type defaultEncoder struct{}

func (e *defaultEncoder) encode(packet *Packet) ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入协议头
	if err := binary.Write(buf, binary.BigEndian, packet.Header); err != nil {
		return nil, err
	}

	// 写入数据
	if _, err := buf.Write(packet.Data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DefaultDecoder 默认解码器
type defaultDecoder struct{}

func (d *defaultDecoder) decode(reader io.Reader) (*Packet, error) {
	// 读取协议头
	header := &packetHeader{}
	if err := binary.Read(reader, binary.BigEndian, header); err != nil {
		return nil, err
	}

	// 验证魔数
	if header.Magic != MagicNumber {
		return nil, fmt.Errorf("invalid magic number")
	}

	// 验证数据长度
	if header.Length > MaxPacketSize {
		return nil, fmt.Errorf("packet too large: %d", header.Length)
	}

	// 读取数据
	data := make([]byte, header.Length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}

	packet := &Packet{
		Header: header,
		Data:   data,
	}

	// 验证数据包
	if !packet.validate() {
		return nil, fmt.Errorf("packet validation failed")
	}

	return packet, nil
}
