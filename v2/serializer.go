package v2

import (
	"encoding/json"
)

// Serializer 序列化接口
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

// JSONSerializer JSON序列化实现
type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JSONSerializer) Deserialize(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewJSONSerializer 创建JSON序列化器
func NewJSONSerializer() Serializer {
	return &JSONSerializer{}
}
