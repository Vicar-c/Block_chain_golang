package network

import (
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestConnect(t *testing.T) {
	tara := NewLocalTransport("a")
	tarb := NewLocalTransport("b")

	tara.Connect(tarb)
	tarb.Connect(tara)
	addrA, _ := tara.peers.Load(tarb.Addr())
	addrB, _ := tarb.peers.Load(tara.Addr())
	//addr_a := addrA.(*LocalTransport)
	//addr_b := addrB.(*LocalTransport)
	assert.Equal(t, addrA, tarb)
	assert.Equal(t, addrB, tara)
}

func TestSendMessage(t *testing.T) {
	tara := NewLocalTransport("a")
	tarb := NewLocalTransport("b")

	tara.Connect(tarb)
	tarb.Connect(tara)

	msg := []byte("hello world")
	assert.Nil(t, tara.SendMessage(tarb.Addr(), msg))

	rpc := <-tarb.Consume()
	payloadBytes, err := io.ReadAll(rpc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, payloadBytes, msg)
	assert.Equal(t, rpc.From, tara.Addr())
}

func TestBroadcast(t *testing.T) {
	tara := NewLocalTransport("a")
	tarb := NewLocalTransport("b")
	tarc := NewLocalTransport("c")
	tara.Connect(tarb)
	tara.Connect(tarc)

	msg := []byte("hello world")
	assert.Nil(t, tara.Broadcast(msg))
	rpcb := <-tarb.Consume()
	b, err := io.ReadAll(rpcb.Payload)
	assert.Nil(t, err)
	assert.Equal(t, b, msg)
	rpcc := <-tarc.Consume()
	c, err := io.ReadAll(rpcc.Payload)
	assert.Nil(t, err)
	assert.Equal(t, c, msg)

}
