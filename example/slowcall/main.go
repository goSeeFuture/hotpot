// package main
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/goSeeFuture/hub"
)

func main() {

	// 目标页面地址
	targets := []string{
		"https://gitee.com/GodanKing/hotpot",
		"https://github.com/PuerkitoBio/goquery",
	}

	// 定义一个通道监听组
	g := hub.NewGroup()
	// 用于等待所有请求返回
	var wg sync.WaitGroup
	for _, link := range targets {
		wg.Add(1)

		// 使用额外的协程执行HTTP请求
		g.SlowCall(func(arg interface{}) (ret hub.Return) {
			link := arg.(string) // 将g.SlowCall第2个参数转为原本的类型
			var resp *http.Response
			fmt.Println("请求：", link)
			resp, ret.Error = http.Get(link)
			if ret.Error != nil {
				return
			}

			fmt.Println("收到", link, "返回的数据")
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}

			// 查找title节点的内容
			title := regexp.MustCompile(`<title>.*</title>`).FindString(string(body))
			if title != "" {
				title = title[7 : len(title)-8]
			}
			ret.Value = title
			return
		}, link, func(ret hub.Return) {
			defer wg.Done()

			if ret.Error != nil {
				fmt.Println("http请求出错：", ret.Error)
				return
			}

			fmt.Println("读出标题\n《", ret.Value.(string), "》")
		})
	}

	wg.Wait()
}
