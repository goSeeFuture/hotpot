package union

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

// NetConn golang默认连接
type NetConn struct {
	conn net.Conn

	msgLenBytes int
	maxMsgLen   int
}

func NewNetConn(conn net.Conn, msgLenBytes, maxMsgLen int) *NetConn {
	return &NetConn{conn: conn, maxMsgLen: maxMsgLen, msgLenBytes: msgLenBytes}
}

func (nc *NetConn) WriteMessage(mt ControlType, data []byte) error {
	var err error
	switch mt {
	case TextMessage:
		err = nc.writeTextMessage(data)
	case BinaryMessage:
		err = nc.writeBinraryMessage(data)
	}

	return err
}

func (nc *NetConn) ReadMessage(mt ControlType) ([]byte, error) {
	var err error
	var data []byte
	switch mt {
	case TextMessage:
		data, err = nc.readTextMessage()
	case BinaryMessage:
		data, err = nc.readBinraryMessage()
	}
	return data, err
}

func (nc *NetConn) Close() error {
	return nc.conn.Close()
}

func (nc *NetConn) LocalAddr() net.Addr {
	return nc.conn.LocalAddr()
}

func (nc *NetConn) RemoteAddr() net.Addr {
	return nc.conn.RemoteAddr()
}

func (nc *NetConn) SetReadDeadline(t time.Time) error {
	return nc.conn.SetReadDeadline(t)
}

func (nc *NetConn) SetWriteDeadline(t time.Time) error {
	return nc.conn.SetWriteDeadline(t)
}

// readBinraryMessage 读二进制消息
func (nc *NetConn) readBinraryMessage() ([]byte, error) {
	lenBuffer := make([]byte, nc.msgLenBytes)
	_, err := io.ReadFull(nc.conn, lenBuffer)
	if err != nil {
		return nil, err
	}

	var msgLen uint32
	switch nc.msgLenBytes {
	case 2:
		msgLen = uint32(binary.BigEndian.Uint16(lenBuffer))
	case 4:
		msgLen = binary.BigEndian.Uint32(lenBuffer)
	default:
		panic("unsupport MsgLenBytes")
	}

	if msgLen > uint32(nc.maxMsgLen) {
		return nil, errors.New("message too long(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	} else if msgLen < minMsgLen {
		return nil, errors.New("message too short(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	}

	msg := make([]byte, msgLen)
	if _, err := io.ReadFull(nc.conn, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// readTextMessage 读文本消息
func (nc *NetConn) readTextMessage() ([]byte, error) {
	reader := bufio.NewReader(nc.conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) == 0 {
		return nil, errors.New("no packet size")
	}

	// 去掉\r
	line = line[:len(line)-1]
	return line, nil
}

func (nc *NetConn) writeTextMessage(msg []byte) error {
	msgLen := len(msg)
	if msgLen > nc.maxMsgLen {
		return errors.New("message too long(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	} else if msgLen < minMsgLen {
		return errors.New("message too short(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	}

	msg = append(msg, textEnd)
	_, err := nc.conn.Write(msg)
	return err
}

// writeBinraryMessage 写二进制消息
func (nc *NetConn) writeBinraryMessage(msg []byte) error {
	var msgLen = uint32(len(msg))

	// check len
	if msgLen > uint32(nc.maxMsgLen) {
		return errors.New("message too long(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	} else if msgLen < minMsgLen {
		return errors.New("message too short(" + strconv.FormatUint(uint64(msgLen), 10) + "/" + strconv.FormatUint(uint64(nc.maxMsgLen), 10) + ")")
	}

	buffer := make([]byte, uint32(nc.msgLenBytes)+msgLen)
	// write len
	switch nc.msgLenBytes {
	case 2:
		binary.BigEndian.PutUint16(buffer, uint16(msgLen))
	case 4:
		binary.BigEndian.PutUint32(buffer, msgLen)
	default:
		panic("unsupport MsgLenBytes")
	}

	// write data
	buffer = append(buffer[:nc.msgLenBytes], msg...)
	_, err := nc.conn.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}
