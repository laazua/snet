package snet

import (
	"encoding/binary"
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
	// buf := new(bytes.Buffer)
	buf := GetBytesBuffer()
	defer PutBytesBuffer(buf)

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
		return nil, ErrMagicNumberInvalid
	}

	// 验证数据长度
	if header.Length > MaxPacketSize {
		return nil, ErrPacketTooLarge
	}

	// 使用字节切片池读取数据
	dataBuf := GetByteSlice()
	defer PutByteSlice(dataBuf)

	// 确保有足够容量
	if cap(dataBuf) < int(header.Length) {
		// 容量不足，创建新的切片
		dataBuf = make([]byte, header.Length)
	} else {
		// 重用切片，设置正确长度
		dataBuf = dataBuf[:header.Length]
	}

	if _, err := io.ReadFull(reader, dataBuf); err != nil {
		return nil, err
	}

	// 创建数据副本（重要！因为池中的切片会被复用）
	data := make([]byte, header.Length)
	copy(data, dataBuf)

	packet := &Packet{
		Header: header,
		Data:   data,
	}

	if !packet.validate() {
		return nil, ErrPacketIvalid
	}

	return packet, nil
}
