package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	I32Mul = addInstruction(opcodes.I32Mul, func(ctx *execution.Context) error {
		b := ctx.Stack.Pop().(int32)
		a := ctx.Stack.Pop().(int32)
		ctx.Stack.Push(a * b)
		return nil
	})
)
