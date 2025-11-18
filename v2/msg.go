package v2

// BaseMessage 基础消息结构
type BaseMessage struct {
	ID      uint32
	DataLen uint32
	Data    []byte
}

func (m *BaseMessage) GetMsgID() uint32 {
	return m.ID
}

func (m *BaseMessage) GetData() []byte {
	return m.Data
}

func (m *BaseMessage) SetData(data []byte) {
	m.Data = data
	m.DataLen = uint32(len(data))
}

func (m *BaseMessage) GetDataLen() uint32 {
	return m.DataLen
}

// NewMessage 创建新消息
func NewMessage(id uint32, data []byte) Message {
	return &BaseMessage{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}
