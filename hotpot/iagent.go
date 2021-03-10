package hotpot

import (
	"github.com/goSeeFuture/hub"
)

// IAgent 客服端连接代理
type IAgent interface {
	// 获取数据（原子操作）
	Data() interface{}
	// 设置数据（原子操作）
	SetData(data interface{})

	// 发送消息到远端
	WriteMsg(interface{})
	// 关闭连接
	Close()
	// 已关闭
	IsClosed() bool

	// 辅助功能
	Help() IHelper

	// 委托其他工作组处理请求消息
	Delegate(g *hub.Group)
	// 取消委托，自己处理请求
	SelfSupport()

	// 获取IP地址
	RemoteIP() string
	// 上次收到请求时间
	LastReceiveTime() int64

	// 唯一标识
	ID() int64

	// 数据处理链
	Processors() *hub.Queue
}
