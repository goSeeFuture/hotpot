package codec

import (
	"encoding/binary"
	"reflect"
	"time"

	mp "github.com/vmihailenco/msgpack/v5"
)

const extJSBuffer = 0x1B
const extJSDate = 0x0D

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

func decJSDate(dec *mp.Decoder, v reflect.Value, extLen int) error {
	buf8 := make([]byte, 8)
	_, err := dec.Buffered().Read(buf8)
	if err != nil {
		return err
	}

	val := binary.BigEndian.Uint64(buf8)
	ptr := v.Addr().Interface().(*time.Time)
	*ptr = time.Unix(int64(val/1000), int64(val)%1000*int64(time.Millisecond))
	return nil
}

func setupJSMessagePack() {
	mp.RegisterExtDecoder(extJSDate, time.Time{}, decJSDate)
	mp.RegisterExt(extJSBuffer, (*JSBuffer)(nil))
}
