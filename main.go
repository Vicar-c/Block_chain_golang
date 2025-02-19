package main

import (
	"block_chain/crypto"
	"block_chain/network"
	"log"
	"time"
)

func main() {
	privKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", &privKey, ":3000", []string{":3001"}, ":9000")
	go localNode.Start()

	remoteNode := makeServer("REMOTE_NODE", nil, ":3001", []string{":3002"}, "")
	go remoteNode.Start()

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, ":3002", []string{}, "")
	go remoteNodeB.Start()

	//go func() {
	//	time.Sleep(11 * time.Second)
	//	lateNode := makeServer("LATE_NODE", nil, ":3003", []string{":3001"}, "")
	//	go lateNode.Start()
	//}()

	time.Sleep(1 * time.Second)
	//tcpTester()
	select {}
}

//func tcpTester() {
//	conn, err := net.Dial("tcp", ":3000")
//	if err != nil {
//		panic(err)
//	}
//	privKey := crypto.GeneratePrivateKey()
//	tx := core.NewTransaction(contract())
//	tx.Sign(privKey)
//	buf := &bytes.Buffer{}
//	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
//		panic(err)
//	}
//	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
//
//	conn.Write(msg.Bytes())
//}

func makeServer(id string, pk *crypto.PrivateKey, addr string, seedNodes []string, apiListenAddr string) *network.Server {
	opts := network.ServerOpts{
		APIListenAddr: apiListenAddr,
		SeedNodes:     seedNodes,
		ListenAddr:    addr,
		ID:            id,
		PrivateKey:    pk,
	}

	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

//func sendTransaction(tr network.Transport, to net.Addr) error {
//	privKey := crypto.GeneratePrivateKey()
//	tx := core.NewTransaction(contract())
//	tx.Sign(privKey)
//	buf := &bytes.Buffer{}
//	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
//		return err
//	}
//	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
//	return tr.SendMessage(to, msg.Bytes())
//}

func contract() []byte {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	return data
}
