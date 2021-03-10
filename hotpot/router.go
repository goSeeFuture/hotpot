package hotpot

import (
	"sync"
)

// Router 消息路由器
type Router struct {
	// 请求消息处理函数
	processors sync.Map
}

type MessageHandle func(data []byte, a IAgent)

// Get 查找消息处理函数
func (r *Router) Get(from string) MessageHandle {
	fn, exist := r.processors.Load(from)
	if !exist {
		return nil
	}
	return fn.(MessageHandle)
}

// Set 设置请求消息处理
func (r *Router) Set(from string, to MessageHandle) {
	r.processors.Store(from, to)
}
