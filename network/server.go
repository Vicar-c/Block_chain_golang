package network

import (
	"block_chain/api"
	"block_chain/core"
	"block_chain/crypto"
	"block_chain/types"
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/go-kit/log"
	"net"
	"os"
	"sync"
	"time"
)

var defaultBlockTime = 5 * time.Second

type ServerOpts struct {
	APIListenAddr string
	SeedNodes     []string
	ListenAddr    string
	ID            string
	Logger        log.Logger
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
	BlockTime     time.Duration
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	// 匿名嵌套结构体，在没有显式初始化ServerOpts的情况下，对内部字段的赋值不会生效，但不会报错（直接就是0值）
	ServerOpts
	TCPTransport *TCPTransport
	mu           sync.RWMutex
	peerMap      map[net.Addr]*TCPPeer
	trans        []net.Addr
	memPool      *TxPool
	chain        *core.Blockchain
	isValidator  bool
	rpcCh        chan RPC
	quitChan     chan struct{}
	txChan       chan *core.Transaction
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == time.Duration(0) {
		opts.BlockTime = defaultBlockTime
	}
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}
	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "addr", opts.ID)
	}

	// new()仅仅只会分配一块零值的内存
	chain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	// Channel being used to communicate between the JSON RPC server
	// and the node that will process this message.
	txChan := make(chan *core.Transaction)

	if len(opts.APIListenAddr) > 0 {
		//fmt.Println(opts.APIListenAddr)
		apiServerConfig := api.ServerConfig{
			Logger:     opts.Logger,
			ListenAddr: opts.APIListenAddr,
		}
		apiServer := api.NewServer(apiServerConfig, chain, txChan)
		go apiServer.Start()
		opts.Logger.Log("msg", "JSON API server running", "port", opts.APIListenAddr)
	}

	peerCh := make(chan *TCPPeer)
	tr := NewTcpTransport(opts.ListenAddr, peerCh)
	s := &Server{
		TCPTransport: tr,
		peerMap:      make(map[net.Addr]*TCPPeer),
		ServerOpts:   opts,
		chain:        chain,
		memPool:      NewTxPool(1000),
		isValidator:  opts.PrivateKey != nil,
		rpcCh:        make(chan RPC),
		quitChan:     make(chan struct{}, 1),
		txChan:       txChan,
	}
	if s.RPCProcessor == nil {
		// 如果RPCProcessor不存在，则调用Server自己的接口实现，即调用自己的ProcessMessage函数
		s.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) bootstrapNetwork() {
	for _, addr := range s.SeedNodes {
		fmt.Println("trying to connect to ", addr)
		go func(addr string) {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Printf("could not connect to %+v\n", conn)
				return
			}
			s.TCPTransport.peerCh <- &TCPPeer{
				conn: conn,
			}
			//fmt.Println("本地端口:", conn.LocalAddr().String())      // 例如 127.0.0.1:57088
			//fmt.Println("远程(发送)端口:", conn.RemoteAddr().String()) // 127.0.0.1:4000
			//fmt.Println("TCP peer ", conn)
		}(addr)
	}
}

func (s *Server) Start() {
	s.TCPTransport.Start()
	time.Sleep(time.Second * 1)
	s.bootstrapNetwork()
	//s.Logger.Log("msg", "accepting TCP connection on", "addr", s.ListenAddr, "id", s.ID)
free:
	for {
		select {
		case peer := <-s.TCPTransport.peerCh:
			s.peerMap[peer.conn.RemoteAddr()] = peer
			//fmt.Println("TCP listen addr is", s.ListenAddr, "remote is", peer.conn.RemoteAddr())
			go peer.readLoop(s.rpcCh)
			if err := s.sendGetStatusMessage(peer); err != nil {
				s.Logger.Log("err", err)
				continue
			}
			//s.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
		case tx := <-s.txChan:
			if err := s.processTransaction(tx); err != nil {
				s.Logger.Log("process TX error", err)
			}
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				s.Logger.Log("error", err)
				continue
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					s.Logger.Log("error", err)
				}
			}
		case <-s.quitChan:
			break free
		}
	}
	s.Logger.Log("msg", "Server is shutting down")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)

	//s.Logger.Log("msg", "Starting validator loop", "blockTime", s.BlockTime)

	for {
		//fmt.Println("creating new block")

		if err := s.createNewBlock(); err != nil {
			s.Logger.Log("create block error", err)
		}

		<-ticker.C
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {

	switch t := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(t)
	case *core.Block:
		//s.Logger.Log("msg", "process block")
		return s.processBlock(t)
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return s.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, t)
	}

	return nil
}

func (s *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	//s.Logger.Log("msg", "received getBlocks message", "from", from)
	var (
		blocks    = []*core.Block{}
		ourHeight = s.chain.Height()
	)

	if data.To == 0 {
		for i := int(data.From); i <= int(ourHeight); i++ {
			block, err := s.chain.GetBlock(uint32(i))
			if err != nil {
				return err
			}

			blocks = append(blocks, block)
		}
	}

	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := NewMessage(MessageTypeBlocks, buf.Bytes())
	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	return peer.Send(msg.Bytes())
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	getStatusMsg := new(GetStatusMessage)
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return nil
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	//fmt.Println("listen server is ", s.listenAddr, " server peerMap is", s.peerMap)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for netAddr, peer := range s.peerMap {
		if err := peer.Send(payload); err != nil {
			return fmt.Errorf("peer send error => error %s [err: %s]", netAddr, err)
		}
		fmt.Println("Server", s.ListenAddr, " broadcast to ", netAddr, " from ", peer)
	}

	return nil
}

func (s *Server) processBlocksMessage(from net.Addr, data *BlocksMessage) error {
	//s.Logger.Log("msg", "received BLOCKS!!!!!!!!", "from", from)

	for _, block := range data.Blocks {
		if err := s.chain.AddBlock(block); err != nil {
			s.Logger.Log("error", err.Error())
			return err
		}
	}

	return nil
}

func (s *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	//s.Logger.Log("msg", "received STATUS message", "from", from)

	if data.CurrentHeight <= s.chain.Height() {
		//s.Logger.Log("msg", "can not sync, blockHeight too low", "ourHeight", s.chain.Height(), "theirHeight", data.CurrentHeight, "height", data.CurrentHeight, "addr", from)
		return nil
	}

	go s.requestBlocksLoop(from)
	return nil
}

func (s *Server) processGetStatusMessage(from net.Addr) error {
	//s.Logger.Log("msg", "received getStatus message", "from", from)

	statusMessage := StatusMessage{
		ID:            s.ID,
		CurrentHeight: s.chain.Height(),
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	return peer.Send(msg.Bytes())
}

func (s *Server) processBlock(b *core.Block) error {
	if err := s.chain.AddBlock(b); err != nil {
		s.Logger.Log("error", err.Error())
		return err
	}
	//fmt.Printf("server listen addr is %s\n", s.listenAddr)
	go s.broadcastBlock(b)

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {

	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	go s.broadcastTx(tx)

	s.memPool.Add(tx)

	return nil
}

// TODO: Find a way to make sure we dont keep syncing when we are at the highest
// block height in the network.
func (s *Server) requestBlocksLoop(peer net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)

	for {
		ourHeight := s.chain.Height()

		//s.Logger.Log("msg", "requesting new blocks", "requesting height", ourHeight+1)

		// In this case we are 100% sure that the node has blocks heigher than us.
		getBlocksMessage := &GetBlocksMessage{
			From: ourHeight + 1,
			To:   0,
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
			return err
		}

		s.mu.RLock()
		defer s.mu.RUnlock()

		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		peer, ok := s.peerMap[peer]
		if !ok {
			return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
		}

		if err := peer.Send(msg.Bytes()); err != nil {
			s.Logger.Log("error", "failed to send to peer", "err", err, "peer", peer)
		}

		<-ticker.C
	}
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return nil
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	// s.Logger.Log("height", s.chain.Height())
	if err != nil {
		return nil
	}

	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err := block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	// 在下一个Block将交易加入后，memPool清空
	s.memPool.ClearPending()

	go s.broadcastBlock(block)

	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:  1,
		DataHash: types.Hash{},
		Height:   0,
		// 这里必须要设为一个单一值,这样才能保证所有的Server初始的状态一致
		// 保证广播时不会因为前面的初始哈希不一致导致广播失败
		Timestamp: 000000,
	}
	coinbase := crypto.PublicKey{}
	b, _ := core.NewBlock(header, nil)
	tx := core.NewTransaction(nil)
	tx.From = coinbase
	tx.To = coinbase
	tx.Value = 10_000_000
	b.Transactions = append(b.Transactions, tx)
	privKey := crypto.GeneratePrivateKey()
	if err := b.Sign(privKey); err != nil {
		panic(err)
	}

	return b
}
