package instructions

import (
	"wasp/wasp/internal/execution"
)

var (
	Call = addInstruction(0x10, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		return nil
	})

	End = addInstruction(0x0b, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})
)
