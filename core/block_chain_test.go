package core

import (
	"block_chain/types"
	"fmt"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(log.NewLogfmtLogger(os.Stderr), randomBlock(t, 0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
	}

	assert.Equal(t, bc.Height(), uint32(lenBlocks))
	assert.Equal(t, len(bc.headers), lenBlocks+1)
}

func TestNewBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
	fmt.Println(bc.Height())
}

func TestHashBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.True(t, bc.HasBlock(0))
}

func TestGetHeader(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	lenBlocks := 100
	for i := 0; i < lenBlocks; i++ {
		block := randomBlock(t, uint32(i+1), getPrevBlockHash(t, bc, uint32(i+1)))
		assert.Nil(t, bc.AddBlock(block))
		header, err := bc.GetHeader(block.Header.Height)
		assert.Nil(t, err)
		assert.Equal(t, header, block.Header)
	}
}

func TestAddBlockToHigh(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	assert.NotNil(t, bc.AddBlock(randomBlock(t, 3, types.Hash{})))
	assert.Nil(t, bc.AddBlock(randomBlock(t, 1, getPrevBlockHash(t, bc, uint32(1)))))
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)
	return BlockHasher{}.Hash(prevHeader)
}
