package hotpot

import (
	"strconv"
	"testing"
	"time"

	"github.com/goSeeFuture/hub"
)

func testTicker(t *testing.T) func() {
	tm := time.Now()
	return func() {
		t.Log("tick", time.Since(tm))
		Global.AfterFunc(time.Second, testTicker(t))
	}
}

func TestGlobalGroup(t *testing.T) {
	t.Run("测试自定义事件触发", func(t *testing.T) {
		const CustomEvent = "自定义事件"
		Global.ListenCall(CustomEvent, func(arg interface{}) hub.Return {
			t.Log(">>>> 执行事件处理函数", arg)

			value := arg.(int)
			return hub.Return{Value: strconv.Itoa(value) + "=>处理完成"}
		})

		waitResult, registered := Global.Call(CustomEvent, 331122)
		if !registered {
			t.Fatalf("`%s`没有注册处理函数", CustomEvent)
		}
		ar := waitResult()
		if ar.Error != nil {
			t.Fatalf("`%s`处理错误 %v", CustomEvent, ar.Error)
		}

		t.Logf("`%s`处理完成，返回：%v", CustomEvent, ar.Value)
	})

	t.Run("单次定时执行", func(t *testing.T) {
		tm := time.Now()
		Global.AfterFunc(time.Second, func() {
			t.Log("一秒后执行", time.Since(tm))
		})

		time.Sleep(2 * time.Second)
	})

	t.Run("定时执行tick", func(t *testing.T) {
		Global.AfterFunc(time.Second, testTicker(t))
		time.Sleep(10 * time.Second)
	})

	t.Run("慢调用", func(t *testing.T) {
		Global.SlowCall(func(v interface{}) hub.Return {
			<-time.After(time.Second)
			return hub.Return{
				Value: strconv.Itoa(v.(int)) + "=> 慢调用",
			}
		}, 123, func(ar hub.Return) {
			t.Logf("慢调用返回 `%v`", ar.Value)
		})

		time.Sleep(time.Second * 2)
	})

	t.Run("停止", func(t *testing.T) {
		Global.Stop()
		time.Sleep(time.Second * 2)
	})
}

func TestMainHubEmit(t *testing.T) {
	Global.ListenEvent("事件", func(arg interface{}) {
		t.Log("参数", arg)
	})

	go func() {
		for i := 0; i < 10; i++ {
			Global.Emit("事件", "非阻塞")
			time.Sleep(time.Second)
		}
	}()

	t.Log("U see see u")

	Global.ListenEvent("事件", func(arg interface{}) {
		t.Log("%参数", arg)
	})

	time.Sleep(time.Second * 10)
}
