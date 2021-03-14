package codec

type Type string

const (
	JSON             Type = "json"
	MessagePack      Type = "msgpack"
	MessagePackExtJS Type = "msgpack#js"
	Protobuf         Type = "proto"
)

// Serializer 序列化接口
type Serializer interface {
	Marshal(msg interface{}) []byte
	MarshalBody(msg interface{}) []byte
	Unmarshal(data []byte, msg interface{}) error
	Unpack(fmt Type, data []byte) (string, []byte)
}

// IWrapper 消息外包装
type IWrapper interface {
	GetType() string
	GetData() []byte
}

type extInfo struct {
	Setup   func()
	Wrapper IWrapper
}

var (
	// _format 序列化格式
	_format = map[Type]Serializer{
		JSON:             &SerialJSON{},
		MessagePack:      &SerialMessagePack{},
		MessagePackExtJS: &SerialMessagePack{},
		Protobuf:         &SerialProtobuf{},
	}

	_ext = map[Type]extInfo{
		MessagePackExtJS: {Setup: setupJSMessagePack, Wrapper: &MessagePackJSWrapper{}},
	}
)

// Register 注册协议
func Register(fmt Type, codec Serializer) {
	_format[fmt] = codec
}

// GetFormat 获取序列化格式
func Get(fmt Type) Serializer {
	InitExt(fmt)
	return _format[fmt]
}

// InitExt 初始化扩展类型
// 返回格式化后的fmt
func InitExt(fmt Type) {
	ext, exist := _ext[fmt]
	if !exist {
		return
	}

	ext.Setup()
}
