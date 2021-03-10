package network

import (
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/union"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

var (
	agentsCounter int64
)

// AgentMgr 服务器基础类
type AgentMgr struct {
	clients sync.Map

	hRecv  func([]byte) []byte
	hSend  func([]byte, bool) []byte
	hError func(error) union.Conn

	wgAgent sync.WaitGroup
	sync.Mutex

	// 运行（监听）端口
	isruning atomic.Value
}

// NewAgentMgr 新服务器
func newAgentMgr() *AgentMgr {
	am := &AgentMgr{}
	am.isruning.Store(false)
	return am
}

// Shutdown 关服
func (am *AgentMgr) Shutdown(immediately bool) {
	log.Info().Int("agents", am.Len()).Msg("shutdown server...")

	if immediately {
		am.isruning.Store(false)
		am.clients.Range(func(key, val interface{}) bool {
			if val == nil {
				return true
			}

			val.(hotpot.IAgent).Close()
			return true
		})
		return
	}

	am.clients.Range(func(key, val interface{}) bool {
		if val == nil {
			return true
		}

		val.(hotpot.IAgent).Close()
		return true
	})
	am.isruning.Store(false)

	log.Info().Msg("shutdown server over")
}

// IsServerRuning 查询服务器是否启动（监听）
func (am *AgentMgr) IsServerRuning() bool {
	return am.isruning.Load().(bool)
}

// OnReceived 收到消息时回调
func (am *AgentMgr) OnReceived(msg []byte) []byte {
	if am.hRecv != nil {
		return am.hRecv(msg)
	}
	return msg
}

// OnSend 发送消息前回调
func (am *AgentMgr) OnSend(msg []byte, keepalived bool) []byte {
	if am.hSend != nil {
		return am.hSend(msg, keepalived)
	}
	return msg
}

// OnError 错误时回调
func (am *AgentMgr) OnError(err error) union.Conn {
	if am.hError != nil {
		return am.hError(err)
	}
	return nil
}

// Error 连接错误时回调
func (am *AgentMgr) Error(cb func(err error) union.Conn) {
	am.hError = cb
}

// Received 收到消息时回调
func (am *AgentMgr) Received(cb func(msg []byte) []byte) {
	am.hRecv = cb
}

// Send 发送消息前回调
func (am *AgentMgr) Send(cb func(msg []byte, keepalived bool) []byte) {
	am.hSend = cb
}

// Agents 遍历全服Agent
func (am *AgentMgr) Agents() []hotpot.IAgent {
	agents := make([]hotpot.IAgent, 0, am.Len())
	am.clients.Range(func(k, v interface{}) bool {
		agents = append(agents, v.(hotpot.IAgent))
		return true
	})
	return agents
}

// AddAgent 增加Agent
func (am *AgentMgr) AddAgent(a hotpot.IAgent) {
	am.wgAgent.Add(1)
	am.clients.Store(a.ID(), a)
	// log.Std.Debug("agent add")
}

// DelAgent 移除Agent
func (am *AgentMgr) DelAgent(a hotpot.IAgent) {
	am.wgAgent.Done()
	am.clients.Delete(a.ID())
	// log.Std.Debug("agent done")
}

// Len 管理的Agent数量
func (am *AgentMgr) Len() (count int) {
	am.clients.Range(func(key interface{}, val interface{}) bool {
		count++
		return true
	})
	return
}

// Get 获取agent
func (am *AgentMgr) Get(id int64) hotpot.IAgent {
	a, ok := am.clients.Load(id)
	if !ok {
		return nil
	}

	return a.(hotpot.IAgent)
}
