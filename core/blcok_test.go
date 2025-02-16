package core

import (
	"block_chain/crypto"
	"block_chain/types"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func randomBlock(t *testing.T, height uint32, prevBlockHash types.Hash) *Block {
	privKey := crypto.GeneratePrivateKey()
	tx := randomTxWithSignature(t)
	header := &Header{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
	}

	b, _ := NewBlock(header, []*Transaction{tx})
	dataHash, err := CalcuateDataHash(b.Transactions)
	assert.Nil(t, err)
	b.Header.DataHash = dataHash
	assert.Nil(t, b.Sign(privKey))
	return b
}

func TestSignBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
}

func TestVerifyBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(t, 0, types.Hash{})
	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()
	assert.NotNil(t, b.Verify())
	b.Header.Height = 100
	assert.NotNil(t, b.Verify())
}

func TestDecodeEncodeBlock(t *testing.T) {
	b := randomBlock(t, 1, types.Hash{})
	buf := &bytes.Buffer{}
	assert.Nil(t, b.Encode(NewGobBlockEncoder(buf)))
	bDecode := new(Block)
	assert.Nil(t, bDecode.Decode(NewGobBlockDecoder(buf)))
	assert.Equal(t, b, bDecode)
}
