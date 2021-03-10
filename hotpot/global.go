package hotpot

import (
	"os"

	"github.com/goSeeFuture/hub"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// 全局组
	Global *hub.Group
	// 管理agent请求处理函数（全局）
	Route Router

	unamegroup int64
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	Global = hub.NewGroup(hub.GroupName("Global"))
}
