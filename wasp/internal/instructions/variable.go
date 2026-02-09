package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	LocalGet = addInstruction(opcodes.LocalGet, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		ctx.Stack.Push(ctx.Local[localIndex].Value)
		return nil
	})
)
