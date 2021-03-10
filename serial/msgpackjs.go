package serial

import (
	mp "github.com/vmihailenco/msgpack"
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
func (w MessagePackJSWrapper) GetData() []byte { return w.Data.Data }

// JSBuffer JavaScript的Buffer类型
type JSBuffer struct {
	Data []byte
}

// MarshalMsgpack 序列化JSBuffer
func (c *JSBuffer) MarshalMsgpack() ([]byte, error) {
	return mp.Marshal(c)
}

// UnmarshalMsgpack 反序列化JSBuffer
func (c *JSBuffer) UnmarshalMsgpack(b []byte) error {
	c.Data = b
	return nil
}

func setupJSMessagePack() {
	mp.RegisterExt(extJSBuffer, (*JSBuffer)(nil))
}
