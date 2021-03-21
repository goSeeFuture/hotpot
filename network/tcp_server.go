package network

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/union"

	"github.com/rs/zerolog/log"
)

// TCPServer 服务器
type TCPServer struct {
	*AgentMgr
	shutdown func()

	config     serverconfig
	serializer codec.Serializer
}

// 创建服务器对象
func newTCPServer(config serverconfig) hotpot.IAgentMgr {
	s := &TCPServer{
		config:     config,
		AgentMgr:   newAgentMgr(),
		serializer: codec.Get(config.Serialize),
	}

	// 注册PONG消息处理
	setupKeepAlived(s)
	setupOfflineDetect(s, config.Keepalived)
	return s
}

// Start 启动服务器
func (ts *TCPServer) Start() {
	go func() {
		tcpAddr, err := net.ResolveTCPAddr(ts.config.Schema, ts.config.Host)
		if err != nil {
			log.Error().Err(err).Str("addr", ts.config.Listen).Msg("resolve addr")
			return
		}

		l, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Error().Err(err).Str("addr", ts.config.Listen).Msg("listen addr")
			return
		}

		ts.isruning.Store(true)
		ts.shutdown = func() {
			if !ts.IsServerRuning() {
				return
			}

			ts.isruning.Store(false)
			l.Close()
		}

		for ts.IsServerRuning() {
			conn, err := l.Accept()
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok && netErr.Temporary() {
					time.Sleep(5 * time.Millisecond)
					continue
				}

				log.Error().Err(err).Str("addr", ts.config.Listen).Msg("accept")
				break
			}

			if ts.Len() > ts.config.MaxConnNumber() {
				log.Warn().Int("max", ts.config.MaxConnNumber()).Msg("server conn max")
				conn.Close()
				continue
			}

			id := atomic.AddInt64(&agentsCounter, 1)
			ts.AddAgent(NewAgent(id, union.NewNetConn(conn, ts.config.MsgLenBytes, ts.config.MaxMsgLen), ts))
		}

		log.Info().Msg("server listen closed")

		l.Close()

		log.Info().Msg("server shutdown")
	}()
}

// Shutdown 停服
func (ts *TCPServer) Shutdown(immediately bool) {
	if immediately {
		ts.shutdown()
	} else {
		for _, a := range ts.Agents() {
			a.SoftClose()
		}
	}
}

// Serializer 消息序列化方式
func (ts *TCPServer) Serializer() codec.Serializer {
	return ts.serializer
}

// Serializer 消息序列化类型
func (ts *TCPServer) SerializeType() codec.Type {
	return ts.config.Serialize
}

// websocket 控制类型
func (ts *TCPServer) WSControlType() union.ControlType {
	return union.BinaryMessage
}

func (ts *TCPServer) SendChanLen() int {
	return ts.config.SendChanLen
}

func (ts *TCPServer) RecvChanLen() int {
	return ts.config.RecvChanLen
}

func (ts *TCPServer) Listen() string { return ts.config.Listen }
