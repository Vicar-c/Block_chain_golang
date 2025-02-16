package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type LocalTransport struct {
	addr      net.Addr
	consumeCh chan RPC
	peers     sync.Map
}

// New还是返回实现结构体，在实现接口时需要严格按照抽象接口类型输入
func NewLocalTransport(addr net.Addr) *LocalTransport {
	return &LocalTransport{
		addr:      addr,
		consumeCh: make(chan RPC, 1024),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumeCh
}

func (t *LocalTransport) Connect(transport Transport) error {
	t.peers.Store(transport.Addr(), transport)
	return nil
}

func (t *LocalTransport) SendMessage(to net.Addr, payload []byte) error {
	// TODO:远端在同步状态时会出现向自己发送消息的问题,并且当同步消息慢时远端可能一直追不上本地创造新块
	if t.Addr() == to {
		return nil
	}
	peer, ok := t.peers.Load(to)
	if !ok {
		return fmt.Errorf("%s: could not send message to: %s", t.Addr(), to)
	}
	// 用新变量名定义并存储，sync.Map存储的Value统一被视为interface，需要额外断言转型
	peerTransport := peer.(*LocalTransport)
	peerTransport.consumeCh <- RPC{
		From:    t.Addr(),
		Payload: bytes.NewReader(payload),
	}
	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	var firstErr error

	t.peers.Range(func(_, value any) bool {
		peerTransport, ok := value.(*LocalTransport)
		if !ok {
			firstErr = fmt.Errorf("invalid peer type")
			return false // 终止遍历
		}
		if err := t.SendMessage(peerTransport.Addr(), payload); err != nil {
			firstErr = err
			return false // 终止遍历
		}
		return true // 继续遍历
	})

	return firstErr
}

func (t *LocalTransport) Addr() net.Addr {
	return t.addr
}
