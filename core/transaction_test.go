package core

import (
	"block_chain/crypto"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignTransaction(t *testing.T) {
	data := []byte("foo")
	tx := &Transaction{Data: data}
	priKey := crypto.GeneratePrivateKey()
	assert.Nil(t, tx.Sign(priKey))
	assert.NotNil(t, tx.Signature)
}

func TestVerifyTransaction(t *testing.T) {
	data := []byte("foo")
	tx := &Transaction{Data: data}
	priKey := crypto.GeneratePrivateKey()
	assert.Nil(t, tx.Sign(priKey))
	assert.Nil(t, tx.Verify())
	otherPrivKey := crypto.GeneratePrivateKey()
	tx.From = otherPrivKey.PublicKey()
	assert.NotNil(t, tx.Verify())

}

func TestTxEncodeDecode(t *testing.T) {
	tx := randomTxWithSignature(t)
	buf := &bytes.Buffer{}
	assert.Nil(t, tx.Encode(NewGobTxEncoder(buf)))

	txDecode := new(Transaction)
	assert.Nil(t, txDecode.Decode(NewGobTxDecoder(buf)))
	assert.Equal(t, tx, txDecode)
}

func randomTxWithSignature(t *testing.T) *Transaction {
	privKey := crypto.GeneratePrivateKey()
	tx := Transaction{Data: []byte("foo")}
	assert.Nil(t, tx.Sign(privKey))
	return &tx
}
