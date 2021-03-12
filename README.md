# hotpot

hotpot（[火锅](https://gitee.com/GodanKing/hotpot)）轻量网络框架。

## 特点

- 支持TCP、Websocket长连接
- 支持JSON、Protocol Buffers、MessagePack编码格式
- 聚合多通道到单协程
- 提供三种常用协程通讯模型

## 简单例子

### echo服务器

```go
package main

import (
    "fmt"
    "github.com/goSeeFuture/hotpot/hotpot"
    "github.com/goSeeFuture/hotpot/network"
    "github.com/goSeeFuture/hotpot/codec"
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
        network.Serialize(codec.JSON),
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
```

启动echo服务器

```bash
go run example/echo/main.go
```

在客户端发送消息

打开 [WebSocket 在线测试](http://www.websocket-test.com/) 输入如下内容

```json
{
    "type": "Echo",
    "data": {
        "Message": "你好，火锅！"
    }
}
```

```md
### 客户端 
连接 ws://localhost:8848
WebSocket 在线测试 收到如下信息

    服务器 0:16:14
    {\type\:\Echo\,\data\:{\Message\:\你好，火锅！\}}
    服务器 0:16:22
    {\type\:\Ping\,\data\:{}}
    服务器 0:16:32
    和服务器断开连接！


### 服务端
Echo服务器打印如下信息

    hotpot 服务器就绪，监听地址 127.0.0.1:8888
    收到客户端消息： 你好，火锅！
    12:16AM INF kill timeout agent agent.data=null agent.id=1 lastRecvTime=2021-02-28T00:16:14+08:00 
    12:16AM ERR read message error="read tcp 127.0.0.1:8888->127.0.0.1:32774: use of closed network connection"
```

可以看出，hotpot服务器收到客户端的Echo消息，并回应了同样消息。

除此之外，客户端还收到了type为Ping的消息，并且在10秒后断开，这其中的原因，保活节会详细说明。

## 保活

英文名Keepalive，TCP、Websocket都是长连接协议，可以通过设置选项开启连接保活。不过hotpot实现基于业务层的Keepalive，原因如下：

- TCP keepalive处于传输层，由操作系统负责，能够判断进程存在，网络通畅，但无法判断进程阻塞或死锁等问题
- 客户端与服务器之间有四层代理或负载均衡，即在传输层之上的代理，只有传输层以上的数据才被转发，例如：nignx反代、socks5等

> 具体请参考  
> [聊聊 TCP 中的 KeepAlive 机制](https://www.sohu.com/a/212429309_355142)  
> [TCP keepalive的详解(解惑)](https://www.cnblogs.com/lanyangsh/p/10926806.html)

hotpot保活采用Ping、Pong实现，具体方法：  

服务端 -> 客户端：发送Ping  
客户端 -> 服务端：客户端收到Ping，立即发送Pong

如何认定连接无响应

服务器每间隔N秒，遍历同客户端的所有连接，检查最后一次收到请求数据时间，如果超过限定时间X秒，则认定该客户端无响应，并主动断开同客户端的连接。

```output
# 客户端
# WebSocket 在线测试 收到如下信息
服务器 0:16:14
{\type\:\Echo\,\data\:{\Message\:\你好，火锅！\}}
服务器 0:16:22
{\type\:\Ping\,\data\:{}} # 服务器检查网络连通的包
服务器 0:16:32
和服务器断开连接！ # 没有收到Pong消息回应，服务器主动断开此客户端

# 服务端
# Echo服务器打印如下信息
hotpot 服务器就绪，监听地址 127.0.0.1:8888
收到客户端消息： 你好，火锅！
# lastRecvTime说明最后一次收到客户端消息是在 0:16:14，由此可以算出客户端已经18秒没有发送消息
12:16AM INF kill timeout agent agent.data=null agent.id=1 lastRecvTime=2021-02-28T00:16:14+08:00 # 服务器结束了超时无响应的连接
12:16AM ERR read message error="read tcp 127.0.0.1:8888->127.0.0.1:32774: use of closed network connection" # 有客户端连接关闭，结合上一条日志，说明这是主动断开客户端连接导致的错误
```

Keepalive 默认值配置在 cfg/config.go 文件中

```golang
// Keepalive 默认配置
am := network.Serve(
    // ...
    // 每隔5秒检查所有连接是否存活
    // 最后收到消息是在 6~12 秒间，则发送 Ping 消息检测存活
    // 最后收到消息超过 12 秒，则认为客户端无响应，主动断开连接
    network.Keepalived(5, 12),
)
```

## 修改保活检测时间

```golang
// 在启动hotpot服务之前执行，比如延长到每分钟检测超时，80秒以上视为无响应
am := network.Serve(
    // ...
    network.Keepalived(60, 80),
)
```

## 鸣谢

特别感谢 [Leaf](https://github.com/name5566/leaf) 库作者，hotpot 项目便是受其启发。
