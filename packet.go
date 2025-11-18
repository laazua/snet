package snet

import "hash/crc32"

// 数据包类型
// const (
// 	PacketTypeData = iota + 1
// 	PacketTypeAck
// 	PacketTypeHeartbeat
// )

// 协议常量
const (
	MagicNumber     = 0x12345678
	ProtocolVersion = 1
	HeaderSize      = 20               // PacketHeader 大小
	MaxPacketSize   = 10 * 1024 * 1024 // 10MB
)

// 数据包协议头
type packetHeader struct {
	Magic    uint32     // 魔数，用于识别协议
	Version  uint8      // 协议版本
	Type     PacketType // 包类型
	Length   uint32     // 数据长度
	Checksum uint32     // CRC32校验和
	Seq      uint32     // 序列号
}

// Packet 数据包结构
type Packet struct {
	Header *packetHeader
	Data   []byte
}

// NewPacket 创建新数据包
func NewPacket(packetType PacketType, data []byte, seq uint32) *Packet {
	header := &packetHeader{
		Magic:   MagicNumber,
		Version: ProtocolVersion,
		Type:    packetType,
		Length:  uint32(len(data)),
		Seq:     seq,
	}
	header.Checksum = calculateChecksum(data)

	return &Packet{
		Header: header,
		Data:   data,
	}
}

// 计算校验和
func calculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// 验证数据包
func (p *Packet) validate() bool {
	if p.Header.Magic != MagicNumber {
		return false
	}
	if p.Header.Version != ProtocolVersion {
		return false
	}
	if p.Header.Length > MaxPacketSize {
		return false
	}
	return p.Header.Checksum == calculateChecksum(p.Data)
}
