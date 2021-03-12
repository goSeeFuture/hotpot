package hotpot

import (
	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/union"
)

// IAgentMgr Agent管理器接口
type IAgentMgr interface {
	Start()
	IsServerRuning() bool
	Shutdown(immediately bool)

	AddAgent(a IAgent)
	DelAgent(a IAgent)
	Get(id int64) IAgent
	Len() int

	Error(cb func(err error) union.Conn)
	// 收到消息时回调
	Received(cb func(msg []byte) []byte)
	// 发送消息前回调
	Send(cb func(msg []byte, keepalived bool) []byte)

	// 管理的所有Agent
	Agents() []IAgent

	Listen() string
	Serializer() codec.Serializer
	SerializeType() codec.Type
}
