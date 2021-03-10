package serial

import (
	"bytes"

	"github.com/rs/zerolog/log"

	mp "github.com/vmihailenco/msgpack"
)

// MessagePackWrapper 消息外包装
type MessagePackWrapper struct {
	Type string
	Data []byte
}

// GetType _
func (w MessagePackWrapper) GetType() string { return w.Type }

// GetData _
func (w MessagePackWrapper) GetData() []byte { return w.Data }

// SerialMessagePack 序列化MessagePack
type SerialMessagePack struct{}

// 强制使用数组模式序列化
func marshalAsArray(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := mp.NewEncoder(&buf).StructAsArray(true).Encode(v)
	return buf.Bytes(), err
}

// Marshal 序列化
func (c SerialMessagePack) Marshal(val interface{}) []byte {
	m, err := marshalAsArray(val)
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack value")
		return nil
	}

	m, err = marshalAsArray(&MessagePackWrapper{
		Type: typeName(val),
		Data: m,
	})
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack warpper")
		return nil
	}

	return m
}

func (c SerialMessagePack) MarshalBody(val interface{}) []byte {
	m, err := marshalAsArray(val)
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack value")
		return nil
	}
	return m
}

// Unmarshal 反序列化
func (c SerialMessagePack) Unmarshal(data []byte, val interface{}) error {
	return mp.Unmarshal(data, val)
}

// Unpack 解开包装
func (c SerialMessagePack) Unpack(format Type, data []byte) (string, []byte) {
	var mw IWrapper = &MessagePackWrapper{}
	ext, exist := _ext[format]
	if exist {
		mw = ext.Wrapper
	}

	err := mp.Unmarshal(data, mw)
	if err != nil {
		log.Error().Err(err).Msg("unmarshal msg warpper")
		return "", nil
	}

	return mw.GetType(), mw.GetData()
}
