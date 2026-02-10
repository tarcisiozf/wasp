package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	Nop = addInstruction(opcodes.Nop, func(ctx *execution.Context) error {
		return nil
	})

	Call = addInstruction(opcodes.Call, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		return nil
	})

	End = addInstruction(opcodes.End, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})
)
