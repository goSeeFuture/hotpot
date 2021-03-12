// main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/network"
)

// Echo协议结构
type Echo struct {
	Message string
}

// Broadcast广播协议
type Broadcast struct {
	Message string
}

func main() {
	// 创建websocket服务器
	am := network.Serve(
		"ws://127.0.0.1:8848",
		network.Serialize(codec.JSON),
		network.TextMsg(true),
		network.Keepalived(60, 80),
	)

	// 路由Echo消息处理函数
	hotpot.Route.Set("Echo", func(data []byte, a hotpot.IAgent) {
		// 解析客户端请求
		var req Echo
		err := am.Serializer().Unmarshal(data, &req)
		if err != nil {
			fmt.Println("wrong proto")
			return
		}

		fmt.Println("收到客户端消息：", req.Message)

		// 返回同样协议内容给客户端
		a.WriteMsg(&req)
	})

	// 每隔5秒钟，全服广播一次服务器当前时间
	hotpot.Global.AfterFunc(time.Second*5, broadcastServerTime(am))

	// 启动服务
	am.Start()
	fmt.Println("github.com/goSeeFuture/hotpot 服务器就绪，监听地址", am.Listen())

	// 等待关服信号，如 Ctrl+C、kill -2、kill -3、kill -15
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}

func broadcastServerTime(am hotpot.IAgentMgr) func() {
	return func() {
		// 全服广播一次服务器当前时间
		s := "当前服务器时间：" + time.Now().Format("2006-01-02 15:04:05")
		for _, a := range am.Agents() {
			a.WriteMsg(&Broadcast{Message: s})
		}

		// 设置下一次定时广播
		hotpot.Global.AfterFunc(time.Second*5, broadcastServerTime(am))
	}
}
