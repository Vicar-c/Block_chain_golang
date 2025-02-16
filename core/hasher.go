package core

import (
	"block_chain/types"
	"crypto/sha256"
)

// 范型接口
type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func (BlockHasher) Hash(h *Header) types.Hash {
	// go的二进制编码库，可以直接用在json和pb中
	// enc := gob.NewEncoder(buf)
	hash := sha256.Sum256(h.Bytes())
	return hash
}

type TxHasher struct {
}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	return sha256.Sum256(tx.Data)
}
