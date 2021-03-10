// main.go
package main

import (
	"fmt"
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/network"
	"github.com/goSeeFuture/hotpot/serial"
	"os"
	"os/signal"
	"syscall"
)

// Echo协议结构
type Echo struct {
	Message string
}

func main() {
	// 创建websocket服务器
	am := network.Serve(
		"ws://127.0.0.1:8848",
		network.Serialize(serial.JSON),
		network.TextMsg(true),
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

	// 启动服务
	am.Start()
	fmt.Println("github.com/goSeeFuture/hotpot 服务器就绪，监听地址", am.Listen())

	// 等待关服信号，如 Ctrl+C、kill -2、kill -3、kill -15
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
