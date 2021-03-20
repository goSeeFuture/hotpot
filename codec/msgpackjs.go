package codec

import (
	"github.com/tinylib/msgp/msgp"
)

const extJSBuffer = 0x1B

// MessagePackJSWrapper 消息外包装
type MessagePackJSWrapper struct {
	Type string
	Data JSBuffer
}

// GetType _
func (w MessagePackJSWrapper) GetType() string { return w.Type }

// GetData _
func (w MessagePackJSWrapper) GetData() []byte { return []byte(w.Data) }

// JSBuffer JavaScript的Buffer类型
type JSBuffer []byte

// MarshalMsgpack 序列化JSBuffer
func (c *JSBuffer) MarshalBinaryTo(b []byte) error {
	copy(b, []byte(*c))
	return nil
}

// UnmarshalMsgpack 反序列化JSBuffer
func (c *JSBuffer) UnmarshalBinary(b []byte) error {
	*c = make([]byte, len(b))
	copy(*c, b)
	return nil
}

func (c *JSBuffer) Len() int {
	return len(*c)
}

func (c *JSBuffer) ExtensionType() int8 {
	return extJSBuffer
}

func setupJSMessagePack() {
	msgp.RegisterExtension(extJSBuffer, func() msgp.Extension { return new(JSBuffer) })
}
