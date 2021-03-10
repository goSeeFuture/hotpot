package union

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSConn gorilla websocket连接
type WSConn struct {
	sync.Mutex
	conn *websocket.Conn
	ip   string
}

type forwardAddr struct {
	address string
}

func (fa forwardAddr) Network() string { return "tcp" }
func (fa forwardAddr) String() string  { return fa.address }

func NewWSConn(conn *websocket.Conn, ip string) *WSConn {
	return &WSConn{conn: conn, ip: ip}
}

func (wsc *WSConn) WriteMessage(ct ControlType, data []byte) error {
	return wsc.conn.WriteMessage(int(ct), data)
}

func (wsc *WSConn) ReadMessage(ct ControlType) ([]byte, error) {
	tp, data, err := wsc.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if tp != int(ct) {
		return nil, errors.New("msg conrol type incorrect, get " + ControlType(tp).String() + ", want " + ct.String() + "")
	}
	return data, err
}

func (wsc *WSConn) Close() error {
	return wsc.conn.Close()
}

func (wsc *WSConn) LocalAddr() net.Addr {
	return wsc.conn.LocalAddr()
}

func (wsc *WSConn) RemoteAddr() net.Addr {
	if wsc.ip != "" {
		return &forwardAddr{address: wsc.ip}
	}
	return wsc.conn.RemoteAddr()
}

func (wsc *WSConn) SetReadDeadline(t time.Time) error {
	return wsc.conn.SetReadDeadline(t)
}

func (wsc *WSConn) SetWriteDeadline(t time.Time) error {
	return wsc.conn.SetWriteDeadline(t)
}
