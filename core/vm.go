package core

import (
	"encoding/binary"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrAdd      Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
	InstrSub      Instruction = 0x0e
	InstrStore    Instruction = 0x0f
	InstrGet      Instruction = 0xae
	InstrMul      Instruction = 0xea
	InstrDiv      Instruction = 0xfd
)

type Stack struct {
	data []any
	// 下一个位置
	sp int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		//sp:   0,
	}
}

func (s *Stack) Push(v any) {
	// ...（展开运算符）用于将切片的元素逐个传递，而不是作为一个整体传递
	s.data = append([]any{v}, s.data...)
	//s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	// 将第一个值移除
	s.data = append(s.data[:0], s.data[1:]...)
	//s.sp--
	return value
}

// VM 虚拟机的本质运行自定义的指令（智能合约代码），对交易数据进行状态转换。
// 这种转换的核心目的是改变区块链的全局状态，确保去中心化的计算能够按规则执行，并且结果不可篡改。
type VM struct {
	data          []byte
	ip            int // instruction pointer
	stack         *Stack
	contractState *State
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		contractState: contractState,
		data:          data,
		ip:            0,
		stack:         NewStack(120),
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])

		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.ip++
		if vm.ip > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrGet:
		var (
			key = vm.stack.Pop().([]byte)
		)
		value, err := vm.contractState.Get(key)
		if err != nil {
			return err
		}
		vm.stack.Push(value)
	case InstrStore:
		var (
			key            = vm.stack.Pop().([]byte)
			value          = vm.stack.Pop()
			serializeValue []byte
		)

		switch v := value.(type) {
		case int:
			serializeValue = serializeInt64(int64(v))
		default:
			panic("TODO: unknown type")
		}
		vm.contractState.Put(key, serializeValue)
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrPushByte:
		vm.stack.Push(vm.data[vm.ip-1])
	case InstrPack:
		// 将指定数量的字节打包为一个byte切片
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		//fmt.Println(c)
		vm.stack.Push(c)
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	case InstrMul:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a * b
		vm.stack.Push(c)
	case InstrDiv:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a / b
		vm.stack.Push(c)
	}
	return nil
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

func deSerializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
