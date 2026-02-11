package instructions

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

type number interface {
	~int32
}

func mul[T number](ctx *execution.Context) error {
	b := castNumber[T](ctx.Stack.Pop())
	a := castNumber[T](ctx.Stack.Pop())
	ctx.Stack.Push(a * b)
	return nil
}

func add[T number](ctx *execution.Context) error {
	b := castNumber[T](ctx.Stack.Pop())
	a := castNumber[T](ctx.Stack.Pop())
	ctx.Stack.Push(a + b)
	return nil
}

func and[T number](ctx *execution.Context) error {
	b := castNumber[T](ctx.Stack.Pop())
	a := castNumber[T](ctx.Stack.Pop())
	ctx.Stack.Push(a & b)
	return nil
}

func eq[T number](ctx *execution.Context) error {
	b := castNumber[T](ctx.Stack.Pop())
	a := castNumber[T](ctx.Stack.Pop())
	if a == b {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func castNumber[T number](item any) T {
	if value, ok := item.(T); ok {
		return value
	}
	if value, ok := item.(int); ok {
		return T(value)
	}
	panic(fmt.Sprintf("expected number, got %T", item))
}

var (
	Const = addInstruction(opcodes.Const, func(ctx *execution.Context) error {
		value := ctx.Body.Varint()
		ctx.Stack.Push(value)
		return nil
	})

	EqI32 = addInstruction(opcodes.EqI32, eq[int32])

	I32Add = addInstruction(opcodes.I32Add, add[int32])
	I32Mul = addInstruction(opcodes.I32Mul, mul[int32])
	I32And = addInstruction(opcodes.I32And, and[int32])
)
