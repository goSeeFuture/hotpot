package codec

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
	"github.com/tinylib/msgp/msgp"
)

//go:generate msgp
//msgp:tuple MessagePackWrapper

// MessagePackWrapper 消息外包装
type MessagePackWrapper struct {
	Type string
	Data JSBuffer `msg:",extension"`
}

// GetType _
func (w MessagePackWrapper) GetType() string { return w.Type }

// GetData _
func (w MessagePackWrapper) GetData() []byte { return []byte(w.Data) }

// SerialMessagePack 序列化MessagePack
type SerialMessagePack struct{}

// Marshal 序列化
func (c SerialMessagePack) Marshal(val interface{}) []byte {
	mar, ok := val.(msgp.Marshaler)
	if !ok {
		log.Error().Msg("no impl msgp.Marshaler")
		return nil
	}

	m, err := mar.MarshalMsg(nil)
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack value")
		return nil
	}

	warpper := MessagePackWrapper{
		Type: typeName(val),
		Data: JSBuffer(m),
	}
	m, err = warpper.MarshalMsg(nil)
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack warpper")
		return nil
	}

	return m
}

func (c SerialMessagePack) MarshalBody(val interface{}) []byte {
	mar, ok := val.(msgp.Marshaler)
	if !ok {
		log.Error().Str("type", reflect.TypeOf(val).String()).Msg("no impl msgp.Marshaler")
		return nil
	}

	m, err := mar.MarshalMsg(nil)
	if err != nil {
		log.Error().Err(err).Msg("marshal msgpack value")
		return nil
	}
	return m
}

// Unmarshal 反序列化
func (c SerialMessagePack) Unmarshal(data []byte, val interface{}) error {
	umar, ok := val.(msgp.Unmarshaler)
	if !ok {
		log.Error().Msg("no impl msgp.Unmarshaler")
		return nil
	}

	_, err := umar.UnmarshalMsg(data)
	if err != nil {
		log.Error().Err(err).Msg("unmarshal msgpack")
		fmt.Println("data", data)
		return nil
	}

	return nil
}

// Unpack 解开包装
func (c SerialMessagePack) Unpack(format Type, data []byte) (string, []byte) {
	warpper := &MessagePackWrapper{}
	mw := warpper

	_, err := warpper.UnmarshalMsg(data)
	if err != nil {
		log.Error().Err(err).Msg("unmarshal msg warpper")
		fmt.Println("data", data)
		return "", nil
	}

	// fmt.Println("type", mw.GetType(), "data", mw.GetData())
	// fmt.Println("type", warpper.Type, "data", warpper.Data.Data)

	return mw.GetType(), mw.GetData()
}
