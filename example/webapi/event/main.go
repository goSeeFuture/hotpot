// main.go
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func echoServe() {
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

	// 监听`广播`事件
	hotpot.Global.ListenEvent("广播", func(arg interface{}) {
		for _, a := range am.Agents() {
			a.WriteMsg(&Broadcast{Message: arg.(string)})
		}
	})

	// 启动服务
	am.Start()
	fmt.Println("github.com/goSeeFuture/hotpot 服务器就绪，监听地址", am.Listen())
}

func webAPIServe() {
	addr := "127.0.0.1:4000"

	http.HandleFunc("/broadcast", func(w http.ResponseWriter, req *http.Request) {
		msg := req.URL.Query()["msg"]
		if len(msg) == 0 {
			fmt.Fprintf(w, "缺少msg参数\n使用说明 http://%s/broadcast?msg=这是一条广播消息", addr)
			return
		}

		// 向Group发出`广播`事件
		hotpot.Global.Emit("广播", msg[0])
	})

	fmt.Println("web server 就绪，地址 http://" + addr)
	http.ListenAndServe(addr, nil)
}

func main() {
	// 启动echo服务
	echoServe()
	// 启动http服务
	go webAPIServe()

	// 等待关服信号，如 Ctrl+C、kill -2、kill -3、kill -15
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
