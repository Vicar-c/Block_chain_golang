package core

import (
	"errors"
	"fmt"
)

var ErrBlockKnown = errors.New("block already known")

type Validator interface {
	ValidateBlock(*Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{bc: bc}
}

func (v *BlockValidator) ValidateBlock(b *Block) error {
	if v.bc.HasBlock(b.Header.Height) {
		// return fmt.Errorf("chain already contains block (%d) with hash (%s)", b.Header.Height, b.Hash(BlockHasher{}))
		return ErrBlockKnown
	}

	if b.Header.Height != v.bc.Height()+1 {
		return fmt.Errorf("block (%s) with height (%d) is too high => current height (%d)", b.Hash(BlockHasher{}), b.Header.Height, v.bc.Height())
	}

	prevHeader, err := v.bc.GetHeader(b.Header.Height - 1)
	if err != nil {
		return err
	}

	hash := BlockHasher{}.Hash(prevHeader)
	if hash != b.Header.PrevBlockHash {
		return fmt.Errorf("the hash of the previous block (%s) is invalid", b.Header.PrevBlockHash)
	}

	if err := b.Verify(); err != nil {
		return err
	}

	return nil
}
