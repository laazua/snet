package snet

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// DataSerializer 数据序列化器
type DataSerializer struct{}

// NewDataSerializer 创建序列化器
func NewDataSerializer() *DataSerializer {
	return &DataSerializer{}
}

// SerializeMap 序列化map数据
func (ds *DataSerializer) SerializeMap(data map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeMap 反序列化map数据
func (ds *DataSerializer) DeserializeMap(data []byte) (map[string]any, error) {
	var result map[string]any
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// SerializeStruct 序列化结构体
func (ds *DataSerializer) SerializeStruct(data any) ([]byte, error) {
	return json.Marshal(data)
}

// DeserializeStruct 反序列化结构体
func (ds *DataSerializer) DeserializeStruct(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
