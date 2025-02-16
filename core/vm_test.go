package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack(t *testing.T) {
	s := NewStack(120)
	s.Push(1)
	s.Push(2)

	value := s.Pop()
	assert.Equal(t, value, 1)
	value = s.Pop()
	assert.Equal(t, value, 2)
}

func TestVM(t *testing.T) {
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	fmt.Println(data)
	contractState := NewState()
	// 存储的key唯一
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())
	fmt.Println(vm.stack.data)
	fmt.Printf("%+v\n", contractState)
	//valueBytes, err := contractState.Get([]byte("FOO"))
	valueBytes := vm.stack.Pop().([]byte)
	value := deSerializeInt64(valueBytes)
	//assert.Nil(t, err)
	//assert.Equal(t, value, int64(5))
	//result := vm.stack.Pop().(int)
	//assert.Equal(t, -1, result)
	//assert.Equal(t, "FOO", string(result))
	//result := vm.stack.Pop()
	assert.Equal(t, value, int64(5))
}
