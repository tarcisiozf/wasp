package instructions

import (
	"wasp/wasp/internal/execution"
)

var (
	I32Mul = addInstruction(0x6c, func(ctx *execution.Context) error {
		b := ctx.Stack.Pop().(int32)
		a := ctx.Stack.Pop().(int32)
		ctx.Stack.Push(a * b)
		return nil
	})
)
