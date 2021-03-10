package serial

import (
	"github.com/rs/zerolog/log"

	"github.com/golang/protobuf/proto"
)

// SerialProtobuf 序列化protobuf
type SerialProtobuf struct{}

// Marshal 序列化
func (c SerialProtobuf) Marshal(msg interface{}) []byte {
	m, err := proto.Marshal(msg.(proto.Message))
	if err != nil {
		log.Error().Err(err).Msg("marshal protobuf data")
		return nil
	}

	m, err = (&ProtoWarpper{Name: typeName(msg), Data: m}).Marshal()
	if err != nil {
		log.Error().Err(err).Msg("marshal protobuf warpper")
		return nil
	}

	return m
}

func (c SerialProtobuf) MarshalBody(val interface{}) []byte {
	m, err := proto.Marshal(val.(proto.Message))
	if err != nil {
		log.Error().Err(err).Msg("marshal protobuf data")
		return nil
	}
	return m
}

// Unmarshal 反序列化
func (c SerialProtobuf) Unmarshal(data []byte, val interface{}) error {
	return proto.Unmarshal(data, val.(proto.Message))
}

// Unpack 解开包装
func (c SerialProtobuf) Unpack(format Type, data []byte) (string, []byte) {
	t := &ProtoWarpper{}
	err := t.Unmarshal(data)
	if err != nil {
		log.Error().Err(err).Msg("unmarshal protobuf warpper")
		return "", nil
	}

	return t.Name, t.Data
}
