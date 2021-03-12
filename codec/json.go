package codec

import (
	"github.com/rs/zerolog/log"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// JSONWrapper 消息外包装
type JSONWrapper struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`

	Code *int    `json:"code,omitempty"`
	Msg  *string `json:"msg,omitempty"`
}

// SerialJSON 序列化JSON
type SerialJSON struct{}

// Marshal 序列化
func (c SerialJSON) Marshal(val interface{}) []byte {
	var m *JSONWrapper
	switch t := val.(type) {
	case *Error:
		m = &JSONWrapper{Type: typeName(val), Code: &t.Code, Msg: &t.Msg}
	case Error:
		m = &JSONWrapper{Type: typeName(val), Code: &t.Code, Msg: &t.Msg}
	default:
		m = &JSONWrapper{Type: typeName(val), Data: val}
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	return data
}

func (c SerialJSON) MarshalBody(val interface{}) []byte {
	m, err := json.Marshal(val)
	if err != nil {
		log.Error().Err(err).Msg("marshal json value")
		return nil
	}
	return m
}

// Unmarshal 反序列化
func (c SerialJSON) Unmarshal(data []byte, val interface{}) error {
	return json.Unmarshal(data, val)
}

// Unpack 解开包装
func (c SerialJSON) Unpack(_ Type, data []byte) (string, []byte) {
	t := json.Get(data, "type")
	if t.LastError() != nil {
		log.Trace().Err(t.LastError()).Msg("unmarshal msg type warpper")
		return "", nil
	}

	m := json.Get(data, "data")
	if m.LastError() != nil {
		log.Trace().Err(t.LastError()).Msg("unmarshal msg data warpper")
		return "", nil
	}

	return t.ToString(), []byte(m.ToString())
}
