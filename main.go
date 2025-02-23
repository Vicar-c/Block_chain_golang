package main

import (
	"block_chain/core"
	"block_chain/crypto"
	"block_chain/network"
	"block_chain/types"
	"block_chain/util"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	validatorPrivKey := crypto.GeneratePrivateKey()
	localNode := makeServer("LOCAL_NODE", &validatorPrivKey, "127.0.0.1:3000", []string{"127.0.0.1:4000"}, "127.0.0.1:9000")

	remoteNode := makeServer("REMOTE_NODE", nil, "127.0.0.1:4000", []string{"127.0.0.1:6000"}, "")

	remoteNodeB := makeServer("REMOTE_NODE_B", nil, "127.0.0.1:6000", nil, "")
	go localNode.Start()
	go remoteNode.Start()
	go remoteNodeB.Start()

	go func() {
		time.Sleep(11 * time.Second)

		lateNode := makeServer("LATE_NODE", nil, ":8000", []string{":4000"}, "")
		go lateNode.Start()
	}()

	time.Sleep(1 * time.Second)

	//if err := sendTransaction(validatorPrivKey); err != nil {
	//	panic(err)
	//}

	// collectionOwnerPrivKey := crypto.GeneratePrivateKey()
	// collectionHash := createCollectionTx(collectionOwnerPrivKey)

	// txSendTicker := time.NewTicker(1 * time.Second)
	// go func() {
	// 	for i := 0; i < 20; i++ {
	// 		nftMinter(collectionOwnerPrivKey, collectionHash)

	// 		<-txSendTicker.C
	// 	}
	// }()

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

func sendTransaction(privKey crypto.PrivateKey) error {
	toPrivKey := crypto.GeneratePrivateKey()

	tx := core.NewTransaction(nil)
	tx.To = toPrivKey.PublicKey()
	tx.Value = 666

	if err := tx.Sign(privKey); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9000/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)

	return err
}

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

func createCollectionTx(privKey crypto.PrivateKey) types.Hash {
	tx := core.NewTransaction(nil)
	tx.TxInner = core.CollectionTx{
		Fee:      200,
		MetaData: []byte("chicken and egg collection!"),
	}
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9000/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	return tx.Hash(core.TxHasher{})
}

func nftMinter(privKey crypto.PrivateKey, collection types.Hash) {
	metaData := map[string]any{
		"power":  8,
		"health": 100,
		"color":  "green",
		"rare":   "yes",
	}

	metaBuf := new(bytes.Buffer)
	if err := json.NewEncoder(metaBuf).Encode(metaData); err != nil {
		panic(err)
	}

	tx := core.NewTransaction(nil)
	tx.TxInner = core.MintTx{
		Fee:             200,
		NFT:             util.RandomHash(),
		MetaData:        metaBuf.Bytes(),
		Collection:      collection,
		CollectionOwner: privKey.PublicKey(),
	}
	tx.Sign(privKey)

	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9000/tx", buf)
	if err != nil {
		panic(err)
	}

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}

func contract() []byte {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	return data
}
