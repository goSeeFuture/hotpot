package hotpot

import (
	"github.com/goSeeFuture/hub"

	"github.com/rs/zerolog/log"
)

type RequestMessage struct {
	Type  string
	Data  []byte
	Agent IAgent
}

type RouteRequest struct {
	router *Router
}

func (RouteRequest) Name() string {
	return "RouteRequest"
}

func (rh *RouteRequest) OnData(data interface{}) interface{} {

	switch x := data.(type) {
	case RequestMessage:
		router := rh.router
		if x.Type == "Ping" || x.Type == "Pong" {
			router = &Route // ping、pong 消息走全局处理
		}

		fn := router.Get(x.Type)
		if fn == nil {
			log.Warn().Str("type", x.Type).Msg("该消息没有handler")
			return nil // 没有该消息的handler
		}

		fn(x.Data, x.Agent)
	default:
		return data
	}
	return nil
}

// 分发agent收到的客户端请求消息
func RouteRequestMessage(router *Router) hub.IDataProcessor {
	return &RouteRequest{router: router}
}
