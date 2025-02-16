package network

import (
	"block_chain/core"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, p.PendingCount(), 0)
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool(10)
	tx := core.NewTransaction([]byte("foo"))
	p.Add(tx)
	assert.Equal(t, p.PendingCount(), 1)
	_ = core.NewTransaction([]byte("fooo"))

	p.ClearPending()
	assert.Equal(t, p.PendingCount(), 0)
}

func TestSortTransaction(t *testing.T) {
	p := NewTxPool(10)
	txLen := 1000

	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		tx.SetFirstSeen(int64(i * rand.Intn(10000)))
		p.Add(tx)
	}

	assert.Equal(t, txLen, p.PendingCount())

	txx := p.Pending()
	assert.Equal(t, len(txx), txLen)
}
