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

	if localDeclCount != 0 {
		panic("unsupported local declarations")
	}

	local := make([]any, len(args)+localDeclCount)

	for i, arg := range args {
		local[i] = arg
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
