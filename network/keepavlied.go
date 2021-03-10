package network

import (
	"time"

	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/serial"
	"github.com/goSeeFuture/hub"

	"github.com/rs/zerolog/log"
)

// setupKeepAlived 设置Agent保活
func setupKeepAlived(s hotpot.IAgentMgr) {
	hotpot.Route.Set("Ping", func(data []byte, a hotpot.IAgent) {
		a.WriteMsg(&serial.Pong{})
	})
	hotpot.Route.Set("Pong", func(data []byte, a hotpot.IAgent) {})
}

// setupOfflineDetect 设置断线检查
func setupOfflineDetect(s hotpot.IAgentMgr, kp keepalived) {
	// 启动定时测试连接Numb
	hotpot.Global.AfterFunc(time.Duration(kp.Interval)*time.Second, numbAgentDetect(s, kp))
}

// numbAgentDetect 检查并关闭麻木Agent
func numbAgentDetect(s hotpot.IAgentMgr, kp keepalived) func() {
	return func() {
		agents := s.Agents()
		hotpot.Global.SlowCall(func(interface{}) hub.Return {
			now := time.Now().Unix()
			drMin := kp.Interval
			drMax := kp.DetectLimit
			limit := kp.Interval + kp.DetectLimit
			var timeOutAgents []int
			for i, a := range agents {
				diff := int(now - a.LastReceiveTime())
				if diff <= drMin {
					continue
				}

				if diff > drMin && diff <= drMax {
					// 测试连接
					a.WriteMsg(&serial.Ping{})
				} else if diff > limit {
					timeOutAgents = append(timeOutAgents, i)
				}
			}

			for _, i := range timeOutAgents {
				a := agents[i]
				log.Info().Int64("agent.id", a.ID()).Interface("agent.data", a.Data()).Time("lastRecvTime", time.Unix(a.LastReceiveTime(), 0)).Msg("kill timeout agent")
				a.Close()
			}

			if s.IsServerRuning() {
				hotpot.Global.AfterFunc(time.Duration(kp.Interval)*time.Second, numbAgentDetect(s, kp))
			}

			return hub.Return{}
		}, nil, nil)
	}
}
