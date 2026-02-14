package instructions

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

type ints interface {
	~int32 | ~int64
}

type floats interface {
	~float32 | ~float64
}

type number interface {
	ints | floats
}

func mul[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a * b)
	return nil
}

func add[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a + b)
	return nil
}

func sub[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a - b)
	return nil
}

func and[T ints](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a & b)
	return nil
}

func eq[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	if a == b {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func castTypedInt[T number](item any) T {
	if value, ok := item.(T); ok {
		return value
	}
	if value, ok := item.(int); ok {
		return T(value)
	}
	panic(fmt.Sprintf("expected number, got %T", item))
}

func castInt(item any) int {
	switch item.(type) {
	case int:
		return item.(int)
	case int32:
		return int(item.(int32))
	case int64:
		return int(item.(int64))
	}
	panic(fmt.Sprintf("expected number, got %T", item))
}

func intConst[T ints](ctx *execution.Context) error {
	value := ctx.Body.Varint()
	ctx.Stack.Push(T(value))
	return nil
}

var (
	I32Const = addInstruction(opcodes.I32Const, intConst[int32])
	I32Eq    = addInstruction(opcodes.I32Eq, eq[int32])
	I32Add   = addInstruction(opcodes.I32Add, add[int32])
	I32Sub   = addInstruction(opcodes.I32Sub, sub[int32])
	I32Mul   = addInstruction(opcodes.I32Mul, mul[int32])
	I32And   = addInstruction(opcodes.I32And, and[int32])

	I64Const = addInstruction(opcodes.I64Const, intConst[int64])

	F32Const = addInstruction(opcodes.F32Const, func(ctx *execution.Context) error {
		value := ctx.Body.Float32()
		ctx.Stack.Push(value)
		return nil
	})

	F64Const = addInstruction(opcodes.F64Const, func(ctx *execution.Context) error {
		value := ctx.Body.Float64()
		ctx.Stack.Push(value)
		return nil
	})
)
