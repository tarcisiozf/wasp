package instructions

import (
	"wasp/wasp/internal/execution"
)

type number interface {
	~int32
}

func mul[T number](ctx *execution.Context) error {
	b := ctx.Stack.Pop().(T)
	a := ctx.Stack.Pop().(T)
	ctx.Stack.Push(a * b)
	return nil
}

var (
	Const = addInstruction(0x41, func(ctx *execution.Context) error {
		value := ctx.Body.Varint()
		ctx.Stack.Push(value)
		return nil
	})

	I32Mul = addInstruction(0x6c, mul[int32])
)
