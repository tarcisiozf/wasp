package module

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/iterator"
	"wasp/wasp/internal/memory"
)

type FunctionSignature struct {
	params  []int
	results []int
}

type Function struct {
	signature FunctionSignature
	body      *iterator.Iterator
}

func (fn *Function) call(stack *memory.Stack, args []any) ([]any, error) {
	localDeclCount := fn.body.Varint()

	// TODO: whats the order? args then local decls? or local decls then args?
	local := make([]memory.Local, 0, len(args)+localDeclCount)

	for _, arg := range args {
		local = append(local, memory.Local{Value: arg})
	}

	for i := 0; i < localDeclCount; i++ {
		localTypeCount := fn.body.Varint()
		for j := 0; j < localTypeCount; j++ {
			localType := fn.body.Byte()
			isConst := isConstOfType(fn.body, localType)
			value := readValue(fn.body, localType)

			local = append(local, memory.Local{
				Value: value,
				Const: isConst,
			})
		}
	}

	ctx := &execution.Context{
		Stack: stack,
		Local: local,
		Body:  fn.body,
	}

	for !ctx.Done {
		opcode := fn.body.Byte()
		ix := instructions.Instruction(opcode)
		if ix == nil {
			return nil, fmt.Errorf("invalid opcode: 0x%x", opcode)
		}
		if err := ix(ctx); err != nil {
			return nil, fmt.Errorf("failed to execute instruction 0x%x: %w", opcode, err)
		}
	}

	results := make([]any, len(fn.signature.results))
	for i := range fn.signature.results {
		results[i] = stack.Pop()
	}
	return results, nil
}

func readValue(iter *iterator.Iterator, typeCode byte) any {
	switch typeCode {
	case typeI32:
		v := iter.Varint()
		return int32(v)
	default:
		panic(fmt.Sprintf("unsupported type code: 0x%x", typeCode))
	}
}

func isConstOfType(iter *iterator.Iterator, typeCode byte) bool {
	if int(typeCode) <= len(constOfTypes) && iter.Peek() == constOfTypes[typeCode] {
		iter.Next()
		return true
	}
	return false
}
