package network

import (
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goSeeFuture/hotpot/codec"
	"github.com/goSeeFuture/hotpot/hotpot"
	"github.com/goSeeFuture/hotpot/union"
	"github.com/goSeeFuture/hub"

	"github.com/rs/zerolog/log"
)

// 连接读写状态
type conniostate int

const (
	unknown   conniostate = iota
	keep                  // 保持读写
	close                 // 关闭
	onlywrite             // 只写
)

type agent struct {
	conn union.Conn

	// 通道监听组
	g *hub.Group

	mgr iMgr

	// N:1
	recvChan     chan interface{}
	stopRecvChan chan struct{}
	// N:1
	sendChan     chan []byte
	stopSendChan chan struct{}

	// 处理通道
	processChan chan interface{}

	// 委托data给其他Group处理
	delegateChan chan interface{}
	// 委托通道监听组
	delegated atomic.Value

	// 是否已经关闭 bool
	iostate atomic.Value

	// 最近一个消息接收时间 int64
	lastRecvTime atomic.Value

	id int64

	data   atomic.Value
	closed atomic.Value
}

// NewAgent 创建Agent
func NewAgent(id int64, conn union.Conn, mgr iMgr) hotpot.IAgent {
	a := &agent{
		id:           id,
		conn:         conn,
		mgr:          mgr,
		stopSendChan: make(chan struct{}),
		sendChan:     make(chan []byte, mgr.SendChanLen()),
		stopRecvChan: make(chan struct{}),
		recvChan:     make(chan interface{}, mgr.RecvChanLen()),
		processChan:  make(chan interface{}, mgr.RecvChanLen()),
		delegateChan: make(chan interface{}, mgr.RecvChanLen()),
	}
	a.g = hub.NewGroup(hub.GroupName("AgentGroup"), hub.GroupHandles(a))
	log.Trace().Int64("id", id).Msg("new agent:")

	// 初始化接收包时间，建立连接算作一次接收
	a.lastRecvTime.Store(time.Now().Unix())
	a.iostate.Store(keep)
	a.closed.Store(false)
	a.g.Attach(a.recvChan)
	a.g.Attach(a.processChan)
	a.delegated.Store(a.g)

	go a.readMessage()
	go a.delayWrite()

	// 全服通知建立连接
	_ = hotpot.Global.Emit(hotpot.EventAgentOpen, a)

	return a
}

func (a *agent) Name() string {
	return "Agent"
}

func (a *agent) OnData(data interface{}) interface{} {
	switch x := data.(type) {
	case []byte: // 从 a.recvChan 接收原始数据
		// 解析消息头
		typ, data := a.mgr.Serializer().Unpack(a.mgr.SerializeType(), x)
		if typ == "" {
			return nil // 不支持的消息头
		}
		// 包装成RequestMessage对象，投递给 a.processChan 或 a.delegateChan
		if a.IsDelegated() {
			a.delegateChan <- hotpot.RequestMessage{Type: typ, Data: data, Agent: a}
		} else {
			a.processChan <- hotpot.RequestMessage{Type: typ, Data: data, Agent: a}
		}
	case hotpot.RequestMessage:
		fn := hotpot.Route.Get(x.Type)
		if fn == nil {
			log.Warn().Str("type", x.Type).Msg("该消息没有handler")
			return nil // 没有该消息的handler
		}
		fn(x.Data, a)
		return data
	default:
		return data
	}
	return nil
}

// LastReceiveTime 上次收到请求时间
func (a *agent) LastReceiveTime() int64 {
	return a.lastRecvTime.Load().(int64)
}

// RemoteIP 获取IP地址
func (a *agent) RemoteIP() string {
	idx := strings.Index(a.conn.RemoteAddr().String(), ":")
	if idx == -1 {
		return a.conn.RemoteAddr().String()
	}

	host, _, err := net.SplitHostPort(a.conn.RemoteAddr().String())
	if err != nil {
		return ""
	}
	return host
}

func (a *agent) WriteMsg(msg interface{}) {
	data := a.mgr.Serializer().Marshal(msg)
	if data == nil {
		return
	}

	var keepavlied bool
	switch msg.(type) {
	case *codec.Ping, *codec.Pong:
		keepavlied = true
	}

	// 回调发送处理函数
	data = a.mgr.OnSend(data, keepavlied)
	if data == nil {
		log.Error().Msg("invalid send msg")
		return
	}

	select {
	case <-a.stopSendChan:
		return
	default:
	}

	if a.ioState() != close {
		a.sendChan <- data
	}
}

func (a *agent) ioState() conniostate {
	return a.iostate.Load().(conniostate)
}

func (a *agent) Close() {
	if a.IsClosed() {
		log.Trace().Msg("already closed")
		return
	}
	log.Trace().Int64("id", a.id).Int("iostate", int(a.iostate.Load().(conniostate))).Msg("close agen")
	a.closed.Store(true)
	a.iostate.Store(close)

	a.g.Stop()        // 关闭集线器
	a.mgr.DelAgent(a) // 通知Server释放连接
	a.conn.Close()    // 关闭连接

	// 全服广播agent关闭
	_ = hotpot.Global.Emit(hotpot.EventAgentClosed, a.Data())
}

// 关闭读消息，只写，待keepalive超时后主动断开
func (a *agent) SoftClose() {
	a.iostate.Store(onlywrite)
}

func (a *agent) readMessage() {
	defer a.Close()

	var err error
	var data []byte
	for a.ioState() == keep {
		// 放入队列
		select {
		case <-a.stopRecvChan:
			return
		default:
		}

		// 接收消息
		data, err = a.conn.ReadMessage(a.mgr.WSControlType())
		if err != nil {
			log.Error().Err(err).Msg("read message")
			// 回调错误处理
			conn := a.mgr.OnError(err)
			if conn != nil {
				// 恢复连接
				a.conn = conn
				continue
			}
			break
		}

		// 记录请求时间
		a.lastRecvTime.Store(time.Now().Unix())
		// 回调接收处理函数
		data = a.mgr.OnReceived(data)
		if data == nil {
			// log.Std.Debug("received invalid msg", zap.String("data", string(data)))
			continue
		}

		select {
		case <-a.stopRecvChan:
			return
		default:
		}

		if a.ioState() == keep {
			a.recvChan <- data
		}
	}
}

func (a *agent) delayWrite() {
	defer a.Close()
	var err error
	for a.ioState() != close {
		select {
		case <-a.stopSendChan:
			if len(a.stopSendChan) == 0 {
				// 消息发送完毕后才能关闭
				return
			}
		default:
		}

		msg := <-a.sendChan
		err = a.conn.WriteMessage(a.mgr.WSControlType(), msg)
		if err != nil {
			log.Error().Err(err).Msg("write message")
			// 回调错误处理
			conn := a.mgr.OnError(err)
			if conn != nil {
				// 恢复连接
				a.conn = conn
				continue
			}
			break
		}
	}
}

// 委托其他工作组处理生产数据
func (a *agent) Delegate(g *hub.Group) {
	if a.IsDelegated() {
		// 已经建立其他委托关系，必须停止才能再次委托
		panic("already delegate to some hub, must stop it first")
	}

	a.delegated.Store(g) // 设置委托标记
	g.Attach(a.delegateChan)
}

// 中止委托关系，并自己处理工作
func (a *agent) SelfSupport() {
	value := a.delegated.Load()
	if value == a.g {
		return // 没有建立委托
	}

	g := value.(*hub.Group)
	g.DetachCB(a.delegateChan, func() {
		a.delegated.Store(a.g)
	})
}

// 中止委托关系
func (a *agent) StopDelegate() bool {
	value := a.delegated.Load()
	if value == a.g {
		return false // 没有建立委托
	}

	g := value.(*hub.Group)

	g.DetachCB(a.delegateChan, func() {
		a.delegated.Store(a.g)
	})

	return true
}

// 是否已经委托
func (a *agent) IsDelegated() bool {
	return a.delegated.Load() != a.g
}

func (a *agent) ID() int64 {
	return a.id
}

func (a *agent) Help() hotpot.IHelper {
	return a.g
}

// Data 获取数据
func (a *agent) Data() interface{} {
	return a.data.Load()
}

// SetData 设置数据
func (a *agent) SetData(v interface{}) {
	a.data.Store(v)
}

// 已关闭
func (a *agent) IsClosed() bool {
	return a.closed.Load().(bool)
}

// 是否保持读写
func (a *agent) IsKeep() bool {
	return a.iostate.Load().(conniostate) == keep
}

// 是否只写
func (a *agent) IsOnlyWrite() bool {
	return a.iostate.Load().(conniostate) == onlywrite
}

// 数据处理链
func (a *agent) Processors() *hub.Queue {
	return a.g.Processors()
}
