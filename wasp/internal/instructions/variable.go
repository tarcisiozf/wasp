package instructions

import (
	"wasp/wasp/internal/execution"
)

var (
	LocalGet = addInstruction(0x20, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		ctx.Stack.Push(ctx.Local[localIndex].Value)
		return nil
	})
)
