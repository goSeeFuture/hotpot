package union

import (
	"net"
	"time"
)

const (
	// msgLenBytes = 2
	minMsgLen = 1
	textEnd   = '\n'
)

// ControlType 控制类型
type ControlType int

const (
	// TextMessage 文本消息
	TextMessage ControlType = 1
	// BinaryMessage 二进制消息
	BinaryMessage ControlType = 2
)

var (
	ctltype = map[ControlType]string{
		TextMessage:   "text",
		BinaryMessage: "binary",
	}
)

func (ct ControlType) String() string {
	return ctltype[ct]
}

// Conn 统一的连接操作
type Conn interface {
	// WriteMessage writes data to the connection.
	WriteMessage(mt ControlType, data []byte) error

	// ReadMessage reads data from the connection.
	ReadMessage(mt ControlType) ([]byte, error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
}
