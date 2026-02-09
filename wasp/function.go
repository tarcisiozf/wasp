package wasp

import "fmt"

const (
	opcodeLocalGet = 0x20

	opcodeI32Mul = 0x6c

	opcodeEnd = 0x0b
)

type Function struct {
	params  []int
	results []int
	body    *Iterator
}

func (fn *Function) call(stack *Stack, args []any) []any {
	localDeclCount := fn.body.varint()

	if localDeclCount != 0 {
		panic("unsupported local declarations")
	}

	local := make([]any, len(args)+localDeclCount)

	for i, arg := range args {
		local[i] = arg
	}

loop:
	for {
		opcode := fn.body.byte()

		switch opcode {
		case opcodeLocalGet:
			localIndex := fn.body.varint()
			stack.push(local[localIndex])
		case opcodeI32Mul:
			b := stack.pop()
			a := stack.pop()
			result := a.(int32) * b.(int32)
			stack.push(result)
		case opcodeEnd:
			break loop
		default:
			panic(fmt.Sprintf("unsupported opcode: 0x%x", opcode))
		}
	}

	results := make([]any, len(fn.results))
	for i := range fn.results {
		results[i] = stack.pop()
	}
	return results
}
