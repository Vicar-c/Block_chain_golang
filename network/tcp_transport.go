package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// TCPPeer 与远端的连接
type TCPPeer struct {
	conn net.Conn
	//From net.Addr
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 4096)
	for {
		n, err := p.conn.Read(buf)
		if err == io.EOF {
			continue
		}
		if err != nil {
			fmt.Printf("read error: %s", err)
			continue
		}
		//fmt.Println("p.conn.RemoteAddr is ", p.conn.RemoteAddr())
		rpcCh <- RPC{
			From:    p.conn.RemoteAddr(),
			Payload: bytes.NewReader(buf[:n]),
		}
		//fmt.Println(string(msg))
	}
}

type TCPTransport struct {
	peerCh     chan *TCPPeer
	listenAddr string
	listener   net.Listener
}

func NewTcpTransport(addr string, peerCh chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		peerCh:     peerCh,
		listenAddr: addr,
	}
}

func (t *TCPTransport) Start() error {
	In, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listener = In

	go t.acceptLoop()
	//fmt.Println("TCP transport listen to port: ", t.listenAddr)
	return nil
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error from %+v\n", conn)
			continue
		}
		//fmt.Println("TCP transport remote port: ", conn.RemoteAddr())
		peer := &TCPPeer{conn: conn}
		t.peerCh <- peer
	}
}
