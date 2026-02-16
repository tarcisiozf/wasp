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

func eqz[T ints](ctx *execution.Context) error {
	a := castTypedInt[T](ctx.Stack.Pop())
	if a == 0 {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func ne[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	if a != b {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func xor[T ints](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a ^ b)
	return nil
}

func or[T ints](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a | b)
	return nil
}

func div[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	ctx.Stack.Push(a / b)
	return nil
}

func gt[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	if a > b {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func lt[T number](ctx *execution.Context) error {
	b := castTypedInt[T](ctx.Stack.Pop())
	a := castTypedInt[T](ctx.Stack.Pop())
	if a < b {
		ctx.Stack.Push(1)
	} else {
		ctx.Stack.Push(0)
	}
	return nil
}

func abs[T floats](ctx *execution.Context) error {
	a := castTypedInt[T](ctx.Stack.Pop())
	if a < 0 {
		ctx.Stack.Push(-a)
	} else {
		ctx.Stack.Push(a)
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
	I32Eqz   = addInstruction(opcodes.I32Eqz, eqz[int32])
	I32Ne    = addInstruction(opcodes.I32Ne, ne[int32])
	I32Xor   = addInstruction(opcodes.I32Xor, xor[int32])
	I32Or    = addInstruction(opcodes.I32Or, or[int32])
	I32Div   = addInstruction(opcodes.I32DivS, div[int32])

	I64Const = addInstruction(opcodes.I64Const, intConst[int64])
	I64Add   = addInstruction(opcodes.I64Add, add[int64])
	I64Sub   = addInstruction(opcodes.I64Sub, sub[int64])
	I64Mul   = addInstruction(opcodes.I64Mul, mul[int64])
	I64Eqz   = addInstruction(opcodes.I64Eqz, eqz[int64])
	I64Eq    = addInstruction(opcodes.I64Eq, eq[int64])
	I64Ne    = addInstruction(opcodes.I64Ne, ne[int64])
	I64And   = addInstruction(opcodes.I64And, and[int64])
	I64Xor   = addInstruction(opcodes.I64Xor, xor[int64])
	I64Or    = addInstruction(opcodes.I64Or, or[int64])
	I64Div   = addInstruction(opcodes.I64DivS, div[int64])

	F32Add = addInstruction(opcodes.F32Add, add[float32])
	F32Sub = addInstruction(opcodes.F32Sub, sub[float32])
	F32Mul = addInstruction(opcodes.F32Mul, mul[float32])
	F32Eq  = addInstruction(opcodes.F32Eq, eq[float32])
	F32Ne  = addInstruction(opcodes.F32Ne, ne[float32])
	F32Div = addInstruction(opcodes.F32Div, div[float32])
	F32Gt  = addInstruction(opcodes.F32Gt, gt[float32])
	F32Lt  = addInstruction(opcodes.F32Lt, lt[float32])
	F32Abs = addInstruction(opcodes.F32Abs, abs[float32])

	F64Add = addInstruction(opcodes.F64Add, add[float64])
	F64Sub = addInstruction(opcodes.F64Sub, sub[float64])
	F64Mul = addInstruction(opcodes.F64Mul, mul[float64])
	F64Eq  = addInstruction(opcodes.F64Eq, eq[float64])
	F64Ne  = addInstruction(opcodes.F64Ne, ne[float64])
	F64Div = addInstruction(opcodes.F64Div, div[float64])
	F64Gt  = addInstruction(opcodes.F64Gt, gt[float64])
	F64Lt  = addInstruction(opcodes.F64Lt, lt[float64])
	F64Abs = addInstruction(opcodes.F64Abs, abs[float64])

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
