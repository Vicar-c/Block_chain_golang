package crypto

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratePrivate(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	// address := pubKey.Address()

	msg := []byte("hello world")
	sig, err := privKey.Sign(msg)
	assert.Nil(t, err)

	assert.True(t, sig.Verify(pubKey, msg))

	fmt.Println(sig)
}

func TestKeypairSignVerifyFail(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	msg := []byte("hello world")
	sig, _ := privKey.Sign(msg)
	otherprivKey := GeneratePrivateKey()
	otherpubKey := otherprivKey.PublicKey()
	assert.False(t, sig.Verify(otherpubKey, msg))
	assert.False(t, sig.Verify(pubKey, []byte("xxx")))

}
