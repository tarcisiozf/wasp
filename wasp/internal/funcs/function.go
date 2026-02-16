package funcs

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/opcodes"
)

type Function struct {
	Signature fnsig.Signature
	Locals    []any
	Body      []byte

	// Blocks maps block start positions to their precomputed targets
	Blocks map[int]fnblock.Target

	Index  int
	Offset int
}

func (f *Function) Call(ctx *execution.Context) error {
	for !ctx.Done {
		opcode := ctx.Body.Opcode()
		ix := instructions.Instruction(opcode)
		//fmt.Printf("Executing instruction %s\n", ix.String())
		if ix.Handler == nil { // TODO: remove before flight
			return fmt.Errorf("unimplemented instruction: %s (0x%x)", opcodes.Name(opcode), opcode)
		}
		if err := ix.Handler(ctx); err != nil {
			return fmt.Errorf("error executing instruction %s: %w", ix.String(), err)
		}
		if ctx.FunctionCallRequest >= 0 {
			return nil // pause execution to handle function call
		}
	}
	return nil
}
