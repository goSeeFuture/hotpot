package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goSeeFuture/hub"
)

// 打印数据处理器
type PrintProcessor struct{}

func (PrintProcessor) Name() string {
	return "Print"
}

func (PrintProcessor) OnData(data interface{}) interface{} {
	fmt.Println("final:", data)
	return nil // 次处已经处理完data，不再向后传递
	// return data // 次处已经处理完data，交给队列后边的处理器，继续处理
}

// 翻倍数据处理器
type TimesProcessor struct{}

func (TimesProcessor) Name() string {
	return "Times×2"
}

func (TimesProcessor) OnData(data interface{}) interface{} {
	data = data.(int) * 2
	fmt.Println("Times×2:", data)
	return data // 次处已经处理完data，交给队列后边的处理器，继续处理
	// return nil // 次处已经处理完data，不再向后传递
}

// 过滤据处理器
type FilterProcessor struct{}

func (FilterProcessor) Name() string {
	return "Filter"
}

func (FilterProcessor) OnData(data interface{}) interface{} {
	if data.(int) == 2 {
		return nil // 丢弃掉2
	}

	fmt.Println("Filter:", data)
	return data // 次处已经处理完data，交给队列后边的处理器，继续处理
	// return nil // 次处已经处理完data，不再向后传递
}

func main() {

	// 构建一个组，并设定自定义数据处理函数
	g := hub.NewGroup(hub.GroupHandles(&TimesProcessor{}, &PrintProcessor{}))

	// 聚合chan通道
	// 交由Processor.OnData处理
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	ch3 := make(chan interface{})
	g.Attach(ch1)
	g.Attach(ch2)
	g.Attach(ch3)

	fmt.Println("g.Processors：", g.Processors().String())
	g.Processors().Insert("Times×2", &FilterProcessor{})
	fmt.Println("g.Processors：", g.Processors().String())

	// 向chan中写入数据
	go func() { ch1 <- 1 }()
	go func() { ch2 <- 2 }()
	go func() { ch3 <- 3 }()

	// wait finished
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-ch
}
