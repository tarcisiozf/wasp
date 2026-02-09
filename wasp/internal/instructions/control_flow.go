package instructions

import (
	"wasp/wasp/internal/execution"
)

var (
	End = addInstruction(0x0b, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})
)
