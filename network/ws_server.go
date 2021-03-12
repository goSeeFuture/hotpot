package network

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/union"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/websocket"
)

// WSServer websocket服务
type WSServer struct {
	*AgentMgr

	shutdown func()
	upgrader websocket.Upgrader

	config     serverconfig
	serializer codec.Serializer
}

// 创建websocket服务器
func newWSServer(config serverconfig) hotpot.IAgentMgr {
	httpTimeOut := time.Duration(config.HTTPTimeout) * time.Second
	wss := &WSServer{
		config: config,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: httpTimeOut,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
		AgentMgr:   newAgentMgr(),
		serializer: codec.Get(config.Serialize),
	}

	// 注册PONG消息处理
	setupKeepAlived(wss)
	setupOfflineDetect(wss, config.Keepalived)

	return wss
}

// Start 启动
func (wss *WSServer) Start() {
	go func() {
		httpTimeOut := time.Duration(wss.config.HTTPTimeout) * time.Second
		ln, err := net.Listen(wss.config.Schema, wss.config.Host)
		if err != nil {
			log.Fatal().Err(err).Msg("listen")
			return
		}

		if wss.config.CertFile != "" || wss.config.KeyFile != "" {
			config := &tls.Config{}
			config.NextProtos = []string{"http/1.1"}

			var err error
			config.Certificates = make([]tls.Certificate, 1)
			config.Certificates[0], err = tls.LoadX509KeyPair(wss.config.CertFile, wss.config.KeyFile)
			if err != nil {
				log.Fatal().Err(err).Msg("load x509 key pair")
			}

			ln = tls.NewListener(ln, config)
		}

		wss.isruning.Store(true)
		wss.shutdown = func() {
			if !wss.IsServerRuning() {
				return
			}

			wss.isruning.Store(false)
			ln.Close()
		}
		httpServer := &http.Server{
			Handler:        wss,
			ReadTimeout:    httpTimeOut,
			WriteTimeout:   httpTimeOut,
			MaxHeaderBytes: 1024,
		}

		err = httpServer.Serve(ln)
		if err != nil {
			log.Fatal().Err(err).Msg("server")
		}
	}()
}

func (wss *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := wss.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("ws upgrade error")
		return
	}

	if wss.Len() > wss.config.MaxConnNumber() {
		log.Warn().Int("cur", wss.Len()).Int("max", wss.config.MaxConnNumber()).Msg("server conn max")
		conn.Close()
		return
	}

	conn.SetReadLimit(int64(wss.config.MaxMsgLen))

	ip := r.Header.Get("X-FORWARDED-FOR")
	if ip == "" {
		ip = strings.TrimSpace(r.Header.Get("X-REAL-IP"))
	} else {
		ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	}

	id := atomic.AddInt64(&agentsCounter, 1)
	wss.AddAgent(NewAgent(id, union.NewWSConn(conn, ip), wss))
}

// Shutdown 停服
func (wss *WSServer) Shutdown(immediately bool) {
	wss.shutdown()
	wss.Shutdown(immediately)
}

// Serializer 消息序列化方式
func (wss *WSServer) Serializer() codec.Serializer {
	return wss.serializer
}

func (wss *WSServer) SerializeType() codec.Type {
	return wss.config.Serialize
}

// websocket 控制类型
func (wss *WSServer) WSControlType() union.ControlType {
	if wss.config.IsText {
		return union.TextMessage
	}
	return union.BinaryMessage
}

func (wss *WSServer) SendChanLen() int {
	return wss.config.SendChanLen
}

func (wss *WSServer) RecvChanLen() int {
	return wss.config.RecvChanLen
}

func (wss *WSServer) Listen() string { return wss.config.Listen }
