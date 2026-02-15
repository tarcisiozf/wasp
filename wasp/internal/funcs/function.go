package funcs

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/instructions"
)

type Function struct {
	Signature Signature
	Locals    []any
	Body      []byte

	// Blocks maps block start positions to their precomputed targets
	Blocks map[int]fnblock.Target

	Index  int
	Offset int
}

func (f *Function) Call(ctx *execution.Context) error {
	for !ctx.Done && ctx.FunctionCallRequest < 0 {
		opcode := ctx.Body.Byte()
		ix := instructions.Instruction(opcode)
		//fmt.Printf("Executing instruction %s\n", ix.String())
		if ix.Handler == nil { // TODO: remove before flight
			return fmt.Errorf("unimplemented instruction: 0x%x", opcode)
		}
		if err := ix.Handler(ctx); err != nil {
			return fmt.Errorf("error executing instruction %s: %w", ix.String(), err)
		}
	}
	if ctx.Done {
		results := make([]any, ctx.NumResults)
		for i := ctx.NumResults - 1; i >= 0; i-- {
			results[i] = ctx.Stack.Pop()
		}
		ctx.Results = results
	}
	return nil
}
