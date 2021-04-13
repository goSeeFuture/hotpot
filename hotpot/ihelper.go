package hotpot

import (
	"time"

	"github.com/goSeeFuture/hub"
)

// IHelper 助手
type IHelper interface {
	// Emit 发送事件，给 group 中的 handler 处理
	Emit(event string, arg interface{}) (registered bool)
	// ListenEvent 绑定事件处理函数到组
	ListenEvent(event string, handler func(arg interface{}))

	// Call 调用事件，跨协程调用group中的函数
	Call(event string, arg interface{}) (result hub.Return, registered bool)
	// ListenCall 注册调用事件到组
	ListenCall(event string, handler func(arg interface{}) hub.Return)

	// SlowCall 慢调用，用协程执行fn，并将结果送回到
	SlowCall(fn func(interface{}) hub.Return, arg interface{}, callback func(hub.Return))
	// AfterFunc 延时执行，超时后通过 group 协程调用 fn
	AfterFunc(dur time.Duration, fn func())
}
