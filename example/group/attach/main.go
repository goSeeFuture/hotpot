package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goSeeFuture/hub"
)

// 自定义数据处理器
// 处理Group中的聚合数据
type Processor struct{}

func (p Processor) Name() string {
	return "MyProcessor"
}

func (p Processor) OnData(data interface{}) interface{} {
	fmt.Println("recv:", data)
	return nil // 次处已经处理完data，不再向后传递
}

func main() {

	// 构建一个组，并设定自定义数据处理函数
	g := hub.NewGroup(hub.GroupHandles(&Processor{}))

	// 聚合chan通道
	// 交由Processor.OnData处理
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	ch3 := make(chan interface{})
	g.Attach(ch1)
	g.Attach(ch2)
	g.Attach(ch3)

	// 向chan中写入数据
	go func() { ch1 <- 1 }()
	go func() { ch2 <- 2 }()
	go func() { ch3 <- 3 }()

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
