package network

import (
	"net/http"
	"net/url"

	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/hotpot"

	"github.com/rs/zerolog/log"
)

type serverconfig struct {
	// 监听地址
	Listen string
	// 消息序列化格式
	Serialize codec.Type
	// 文本协议，仅websocket支持
	IsText bool

	// 读取超时，仅websocket支持
	ReadTimeOut int
	// HTTP超时，仅websocket支持
	HTTPTimeout int

	// SSL 证书，仅websocket支持
	CertFile string
	// SSL 证书Key，仅websocket支持
	KeyFile string

	// MaxMsgLen 最大消息长度
	MaxMsgLen int
	// 消息头大小
	MsgLenBytes int
	// 接收缓冲
	RecvChanLen int
	// 发送缓冲
	SendChanLen int
	// 事件通道长度
	EventChanLen int
	// AfterFunc、慢调用等
	AsyncCallLen int

	Keepalived keepalived
	// 最大连接数
	MaxConnNumber func() int
	// 主机地址
	Host string
	// 协议
	Schema string
	// 地址路径
	Path string
	// http处理函数
	HTTPHandleFuncs []HTTPHandleFunc
	// http处理对象
	HTTPHandlers []HTTPHandler
}

type HTTPHandleFunc struct {
	Pattern string
	Handle  func(http.ResponseWriter, *http.Request)
}

type HTTPHandler struct {
	Pattern string
	Handler http.Handler
}

// Keepalived 保活检测设定
type keepalived struct {
	// 连接保活检查间隔时间（秒）
	Interval int
	// 在 (DetectInterval, DetectLimit]区间中，发送ping消息测试网络连接
	// > DetectLimit 超过该时间无请求视为离线
	DetectLimit int
}

func Keepalived(interval, limit int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.Keepalived.Interval = interval
		s.Keepalived.DetectLimit = limit
	}
}

func MsgLenBytes(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.MsgLenBytes = value
	}
}

func MaxMsgLen(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.MaxMsgLen = value
	}
}

func HTTPTimeout(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.HTTPTimeout = value
	}
}

func ReadTimeOut(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.ReadTimeOut = value
	}
}

func EventChanLen(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.EventChanLen = value
	}
}

func AsyncCallLen(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.AsyncCallLen = value
	}
}

func MaxConnectNumber(fn func() int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.MaxConnNumber = fn
	}
}

func StaticMaxConnectNumber(value int) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.MaxConnNumber = func() int {
			return value
		}
	}
}

func Listen(addr string) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.Listen = addr
	}
}

// "tcp", "tcp4", "tcp6"
func Schema(schema string) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.Schema = schema
	}
}

func TextMsg(enable bool) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.IsText = enable
	}
}

func Serialize(value codec.Type) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.Serialize = value
	}
}

func SSL(certFile, keyFile string) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.CertFile = certFile
		s.KeyFile = keyFile
	}
}

func HTTPHandleFuncs(handles ...HTTPHandleFunc) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.HTTPHandleFuncs = handles
	}
}

func HTTPHandlers(handles ...HTTPHandler) func(s *serverconfig) {
	return func(s *serverconfig) {
		s.HTTPHandlers = handles
	}
}

// 配置选项
type Option func(*serverconfig)

func Serve(addr string, options ...Option) hotpot.IAgentMgr {
	sc := serverconfig{
		Listen:       "ws://127.0.0.1:8848",
		Schema:       "tcp4",
		MaxMsgLen:    4096,
		MsgLenBytes:  2,
		RecvChanLen:  100,
		SendChanLen:  10,
		EventChanLen: 32,
		AsyncCallLen: 120,
		Keepalived: keepalived{
			Interval:    5,
			DetectLimit: 12,
		},
		MaxConnNumber: func() int {
			return 20000
		},
	}
	if addr != "" {
		sc.Listen = addr
	}

	for _, option := range options {
		option(&sc)
	}

	l, err := url.Parse(addr)
	if err != nil {
		log.Fatal().Err(err).Msg("parse serve address")
	}

	sc.Host = l.Host
	sc.Path = l.Path
	if l.Scheme == "ws" && sc.Path == "" {
		sc.Path = "/"
	}

	var am hotpot.IAgentMgr
	switch l.Scheme {
	case "ws", "http":
		am = newWSServer(sc)
	case "tcp":
		am = newTCPServer(sc)
	}

	return am
}
