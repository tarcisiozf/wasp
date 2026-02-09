package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	End = addInstruction(opcodes.End, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})
)
