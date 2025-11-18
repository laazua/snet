package v2

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"sync"
)

var (
	// ErrInvalidData 无效数据错误
	ErrInvalidData = errors.New("invalid data")
	// ErrUnsupportedType 不支持的类型错误
	ErrUnsupportedType = errors.New("unsupported type")
)

// GobCodec GOB编解码器（带缓冲池优化）
type GobCodec struct {
	bufferPool *sync.Pool
}

// NewGobCodec 创建新的GOB编解码器
func NewGobCodec() Codec {
	return &GobCodec{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (g *GobCodec) Encode(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrInvalidData
	}

	// 从缓冲池获取buffer
	buf := g.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		g.bufferPool.Put(buf)
	}()

	// 创建新的encoder，不使用池化因为encoder有状态
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	// 复制数据，避免返回缓冲池内部的切片
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func (g *GobCodec) Decode(data []byte, v interface{}) error {
	if len(data) == 0 || v == nil {
		return ErrInvalidData
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(v)
}

// JSONCodec JSON编解码器（带缓冲池优化）
type JSONCodec struct {
	bufferPool *sync.Pool
}

// NewJSONCodec 创建新的JSON编解码器
func NewJSONCodec() Codec {
	return &JSONCodec{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (j *JSONCodec) Encode(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrInvalidData
	}

	// 从缓冲池获取buffer
	buf := j.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		j.bufferPool.Put(buf)
	}()

	// 使用json.Encoder进行编码
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false) // 禁用HTML转义，提高性能

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	// 移除末尾的换行符
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	// 复制数据
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func (j *JSONCodec) Decode(data []byte, v interface{}) error {
	if len(data) == 0 || v == nil {
		return ErrInvalidData
	}

	// 使用json.Unmarshal进行解码，性能更好
	return json.Unmarshal(data, v)
}

// FastJSONCodec 快速JSON编解码器（使用json.Marshal/Unmarshal）
type FastJSONCodec struct{}

// NewFastJSONCodec 创建快速JSON编解码器
func NewFastJSONCodec() Codec {
	return &FastJSONCodec{}
}

func (f *FastJSONCodec) Encode(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrInvalidData
	}
	return json.Marshal(v)
}

func (f *FastJSONCodec) Decode(data []byte, v interface{}) error {
	if len(data) == 0 || v == nil {
		return ErrInvalidData
	}
	return json.Unmarshal(data, v)
}

// PooledBufferCodec 带缓冲池的编解码器
type PooledBufferCodec struct {
	bufferPool *sync.Pool
	useGob     bool
}

// NewPooledBufferCodec 创建带缓冲池的编解码器
func NewPooledBufferCodec(useGob bool) Codec {
	return &PooledBufferCodec{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		useGob: useGob,
	}
}

func (p *PooledBufferCodec) Encode(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, ErrInvalidData
	}

	if p.useGob {
		return p.gobEncode(v)
	}
	return p.jsonEncode(v)
}

func (p *PooledBufferCodec) Decode(data []byte, v interface{}) error {
	if len(data) == 0 || v == nil {
		return ErrInvalidData
	}

	if p.useGob {
		return p.gobDecode(data, v)
	}
	return p.jsonDecode(data, v)
}

func (p *PooledBufferCodec) gobEncode(v interface{}) ([]byte, error) {
	buf := p.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufferPool.Put(buf)
	}()

	enc := gob.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

func (p *PooledBufferCodec) gobDecode(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(v)
}

func (p *PooledBufferCodec) jsonEncode(v interface{}) ([]byte, error) {
	buf := p.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufferPool.Put(buf)
	}()

	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}

	// 移除末尾的换行符
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func (p *PooledBufferCodec) jsonDecode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// DefaultCodec 默认编解码器（智能选择最优实现）
type DefaultCodec struct {
	codec  Codec
	useGob bool
}

// DefaultCodecOption 默认编解码器选项
type DefaultCodecOption func(*DefaultCodec)

// WithGobPriority 设置优先使用GOB编码
func WithGobPriority() DefaultCodecOption {
	return func(dc *DefaultCodec) {
		dc.useGob = true
	}
}

// WithCustomCodec 设置自定义编解码器
func WithCustomCodec(codec Codec) DefaultCodecOption {
	return func(dc *DefaultCodec) {
		dc.codec = codec
	}
}

// NewDefaultCodec 创建默认编解码器
func NewDefaultCodec(opts ...DefaultCodecOption) Codec {
	dc := &DefaultCodec{
		useGob: false,
	}

	// 应用选项
	for _, opt := range opts {
		opt(dc)
	}

	// 如果没有设置自定义编解码器，使用默认
	if dc.codec == nil {
		if dc.useGob {
			dc.codec = NewPooledBufferCodec(true)
		} else {
			// 默认使用带缓冲池的JSON编解码器
			dc.codec = NewPooledBufferCodec(false)
		}
	}

	return dc
}

func (d *DefaultCodec) Encode(v interface{}) ([]byte, error) {
	return d.codec.Encode(v)
}

func (d *DefaultCodec) Decode(data []byte, v interface{}) error {
	return d.codec.Decode(data, v)
}

// CodecFactory 编解码器工厂
type CodecFactory struct {
	codecs map[string]Codec
	mu     sync.RWMutex
}

var (
	defaultFactory *CodecFactory
	factoryOnce    sync.Once
)

// GetCodecFactory 获取编解码器工厂单例
func GetCodecFactory() *CodecFactory {
	factoryOnce.Do(func() {
		defaultFactory = &CodecFactory{
			codecs: make(map[string]Codec),
		}
		// 注册默认编解码器
		defaultFactory.RegisterCodec("json", NewJSONCodec())
		defaultFactory.RegisterCodec("fast_json", NewFastJSONCodec())
		defaultFactory.RegisterCodec("gob", NewGobCodec())
		defaultFactory.RegisterCodec("pooled_json", NewPooledBufferCodec(false))
		defaultFactory.RegisterCodec("pooled_gob", NewPooledBufferCodec(true))
		defaultFactory.RegisterCodec("default", NewDefaultCodec())
	})
	return defaultFactory
}

// RegisterCodec 注册编解码器
func (f *CodecFactory) RegisterCodec(name string, codec Codec) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codecs[name] = codec
}

// GetCodec 获取编解码器
func (f *CodecFactory) GetCodec(name string) (Codec, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	codec, exists := f.codecs[name]
	return codec, exists
}

// GetDefaultCodec 获取默认编解码器
func (f *CodecFactory) GetDefaultCodec() Codec {
	if codec, exists := f.GetCodec("default"); exists {
		return codec
	}
	return NewDefaultCodec()
}

// 在init函数中注册常用类型，避免gob编解码错误
func init() {
	// 注册基础类型
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[string]int{})
	gob.Register(map[string]float64{})
	gob.Register(map[int]interface{}{})
	gob.Register(map[int]string{})

	// 注册常用切片类型
	gob.Register([]string{})
	gob.Register([]int{})
	gob.Register([]int64{})
	gob.Register([]float64{})
	gob.Register([]bool{})

	// 注册错误类型
	gob.Register(errors.New(""))
}
