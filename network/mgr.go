package network

import (
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/serial"
	"github.com/goSeeFuture/hotpot/union"
)

// iMgr 从Server下放给每个Agent使用的基础功能
type iMgr interface {
	DelAgent(a hotpot.IAgent)

	// 是否运行
	IsServerRuning() bool
	// 收到消息时回调
	OnReceived(msg []byte) []byte
	// 发送消息前回调
	OnSend(msg []byte, keepalived bool) []byte
	// 错误时回调
	OnError(err error) union.Conn

	// 全服Agent
	Agents() []hotpot.IAgent

	// 消息序列化方式
	Serializer() serial.Serializer
	// 序列化类型
	SerializeType() serial.Type
	// websocket 控制类型
	WSControlType() union.ControlType

	SendChanLen() int
	RecvChanLen() int
}
