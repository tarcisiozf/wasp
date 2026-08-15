package instructions

import (
	"fmt"
	"math"
	"math/bits"
	"unsafe"

	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/memory/stack"
	"github.com/tarcisiozf/wasp/internal/opcodes"
)

type ints interface {
	int32 | int64
}

type floats interface {
	float32 | float64
}

type number interface {
	ints | floats
}

func mul[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a * b)
	return nil
}

func add[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a + b)
	return nil
}

func sub[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a - b)
	return nil
}

func and[T ints](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a & b)
	return nil
}

func boolToInt(b bool) int {
	return int(*(*uint8)(unsafe.Pointer(&b)))
}

func eq[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a == b))
	return nil
}

func eqz[T ints](ctx *execution.Context) error {
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a == 0))
	return nil
}

func ne[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a != b))
	return nil
}

func xor[T ints](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a ^ b)
	return nil
}

func or[T ints](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a | b)
	return nil
}

func div[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(a / b)
	return nil
}

func gt[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a > b))
	return nil
}

func lt[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a < b))
	return nil
}

func le[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a <= b))
	return nil
}

func ge[T number](ctx *execution.Context) error {
	b := stack.Pop[T](ctx.Stack)
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(boolToInt(a >= b))
	return nil
}

func abs[T floats](ctx *execution.Context) error {
	a := stack.Pop[T](ctx.Stack)
	// Use math.Abs to properly handle edge cases like -0.0 and NaN
	ctx.Stack.Push(T(math.Abs(float64(a))))
	return nil
}

// Unsigned type mappings
type unsigned interface {
	~uint32 | ~uint64
}

func divU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(int32(a / b))
	return nil
}

func divU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(int64(a / b))
	return nil
}

func remS32(ctx *execution.Context) error {
	b := stack.Pop[int32](ctx.Stack)
	a := stack.Pop[int32](ctx.Stack)
	ctx.Stack.Push(a % b)
	return nil
}

func remS64(ctx *execution.Context) error {
	b := stack.Pop[int64](ctx.Stack)
	a := stack.Pop[int64](ctx.Stack)
	ctx.Stack.Push(a % b)
	return nil
}

func remU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(int32(a % b))
	return nil
}

func remU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(int64(a % b))
	return nil
}

func ltU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(boolToInt(a < b))
	return nil
}

func ltU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(boolToInt(a < b))
	return nil
}

func gtU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(boolToInt(a > b))
	return nil
}

func gtU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(boolToInt(a > b))
	return nil
}

func leU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(boolToInt(a <= b))
	return nil
}

func leU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(boolToInt(a <= b))
	return nil
}

func geU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(boolToInt(a >= b))
	return nil
}

func geU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(boolToInt(a >= b))
	return nil
}

func clz32(ctx *execution.Context) error {
	aVal := stack.Pop[int32](ctx.Stack)
	a := uint32(aVal)
	ctx.Stack.Push(int32(bits.LeadingZeros32(a)))
	return nil
}

func clz64(ctx *execution.Context) error {
	aVal := stack.Pop[int64](ctx.Stack)
	a := uint64(aVal)
	ctx.Stack.Push(int64(bits.LeadingZeros64(a)))
	return nil
}

func ctz32(ctx *execution.Context) error {
	aVal := stack.Pop[int32](ctx.Stack)
	a := uint32(aVal)
	ctx.Stack.Push(int32(bits.TrailingZeros32(a)))
	return nil
}

func ctz64(ctx *execution.Context) error {
	aVal := stack.Pop[int64](ctx.Stack)
	a := uint64(aVal)
	ctx.Stack.Push(int64(bits.TrailingZeros64(a)))
	return nil
}

func popcnt64(ctx *execution.Context) error {
	aVal := stack.Pop[int64](ctx.Stack)
	ctx.Stack.Push(int64(bits.OnesCount64(uint64(aVal))))
	return nil
}

func shl32(ctx *execution.Context) error {
	b := ctx.Stack.PopEntry()
	a := ctx.Stack.PopEntry()
	c := a.U32() << (b.U32() % 32)
	ctx.Stack.Push(int32(c))
	return nil
}

func shl64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(int64(a << (b % 64)))
	return nil
}

func shrS32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	a := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	ctx.Stack.Push(a >> (b % 32))
	return nil
}

func shrS64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	a := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	ctx.Stack.Push(a >> (b % 64))
	return nil
}

func shrU32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := uint32(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(int32(a >> (b % 32)))
	return nil
}

func shrU64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := uint64(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(int64(a >> (b % 64)))
	return nil
}

func rotl32(ctx *execution.Context) error {
	bVal := stack.Pop[int32](ctx.Stack)
	aVal := stack.Pop[int32](ctx.Stack)
	b := int(bVal)
	a := uint32(aVal)
	ctx.Stack.Push(int32(bits.RotateLeft32(a, b)))
	return nil
}

func rotl64(ctx *execution.Context) error {
	bVal := stack.Pop[int64](ctx.Stack)
	aVal := stack.Pop[int64](ctx.Stack)
	b := int(bVal)
	a := uint64(aVal)
	ctx.Stack.Push(int64(bits.RotateLeft64(a, b)))
	return nil
}

func neg[T floats](ctx *execution.Context) error {
	a := stack.Pop[T](ctx.Stack)
	ctx.Stack.Push(-a)
	return nil
}

var ErrTypeMismatch = fmt.Errorf("type mismatch: expected number")

func castInt(item any) (int, error) {
	switch v := item.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	}
	return 0, fmt.Errorf("%w, got %T", ErrTypeMismatch, item)
}

var (
	I32Const = addInstruction(opcodes.I32Const, func(ctx *execution.Context) error {
		value := ctx.Body.SignedVarint32()
		ctx.Stack.PushInt32(value)
		return nil
	})
	I64Const = addInstruction(opcodes.I64Const, func(ctx *execution.Context) error {
		value := ctx.Body.SignedVarint64()
		ctx.Stack.Push(value)
		return nil
	})
	I32Eq   = addInstruction(opcodes.I32Eq, eq[int32])
	I32Add  = addInstruction(opcodes.I32Add, add[int32])
	I32Sub  = addInstruction(opcodes.I32Sub, sub[int32])
	I32Mul  = addInstruction(opcodes.I32Mul, mul[int32])
	I32And  = addInstruction(opcodes.I32And, and[int32])
	I32Eqz  = addInstruction(opcodes.I32Eqz, eqz[int32])
	I32Ne   = addInstruction(opcodes.I32Ne, ne[int32])
	I32Xor  = addInstruction(opcodes.I32Xor, xor[int32])
	I32Or   = addInstruction(opcodes.I32Or, or[int32])
	I32Div  = addInstruction(opcodes.I32DivS, div[int32])
	I32DivU = addInstruction(opcodes.I32DivU, divU32)
	I32RemS = addInstruction(opcodes.I32RemS, remS32)
	I32RemU = addInstruction(opcodes.I32RemU, remU32)
	I32LtS  = addInstruction(opcodes.I32LtS, lt[int32])
	I32LtU  = addInstruction(opcodes.I32LtU, ltU32)
	I32GtS  = addInstruction(opcodes.I32GtS, gt[int32])
	I32GtU  = addInstruction(opcodes.I32GtU, gtU32)
	I32LeS  = addInstruction(opcodes.I32LeS, le[int32])
	I32LeU  = addInstruction(opcodes.I32LeU, leU32)
	I32GeS  = addInstruction(opcodes.I32GeS, ge[int32])
	I32GeU  = addInstruction(opcodes.I32GeU, geU32)
	I32Clz  = addInstruction(opcodes.I32Clz, clz32)
	I32Ctz  = addInstruction(opcodes.I32Ctz, ctz32)
	I32Shl  = addInstruction(opcodes.I32Shl, shl32)
	I32ShrS = addInstruction(opcodes.I32ShrS, shrS32)
	I32ShrU = addInstruction(opcodes.I32ShrU, shrU32)
	I32Rotl = addInstruction(opcodes.I32Rotl, rotl32)

	// I64Const is defined earlier with SignedVarint64
	I64Add    = addInstruction(opcodes.I64Add, add[int64])
	I64Sub    = addInstruction(opcodes.I64Sub, sub[int64])
	I64Mul    = addInstruction(opcodes.I64Mul, mul[int64])
	I64Eqz    = addInstruction(opcodes.I64Eqz, eqz[int64])
	I64Eq     = addInstruction(opcodes.I64Eq, eq[int64])
	I64Ne     = addInstruction(opcodes.I64Ne, ne[int64])
	I64And    = addInstruction(opcodes.I64And, and[int64])
	I64Xor    = addInstruction(opcodes.I64Xor, xor[int64])
	I64Or     = addInstruction(opcodes.I64Or, or[int64])
	I64Div    = addInstruction(opcodes.I64DivS, div[int64])
	I64DivU   = addInstruction(opcodes.I64DivU, divU64)
	I64RemS   = addInstruction(opcodes.I64RemS, remS64)
	I64RemU   = addInstruction(opcodes.I64RemU, remU64)
	I64LtS    = addInstruction(opcodes.I64LtS, lt[int64])
	I64LtU    = addInstruction(opcodes.I64LtU, ltU64)
	I64GtS    = addInstruction(opcodes.I64GtS, gt[int64])
	I64GtU    = addInstruction(opcodes.I64GtU, gtU64)
	I64LeS    = addInstruction(opcodes.I64LeS, le[int64])
	I64LeU    = addInstruction(opcodes.I64LeU, leU64)
	I64GeS    = addInstruction(opcodes.I64GeS, ge[int64])
	I64GeU    = addInstruction(opcodes.I64GeU, geU64)
	I64Clz    = addInstruction(opcodes.I64Clz, clz64)
	I64Ctz    = addInstruction(opcodes.I64Ctz, ctz64)
	I64Popcnt = addInstruction(opcodes.I64Popcnt, popcnt64)
	I64Shl    = addInstruction(opcodes.I64Shl, shl64)
	I64ShrS   = addInstruction(opcodes.I64ShrS, shrS64)
	I64ShrU   = addInstruction(opcodes.I64ShrU, shrU64)
	I64Rotl   = addInstruction(opcodes.I64Rotl, rotl64)

	F32Add = addInstruction(opcodes.F32Add, add[float32])
	F32Sub = addInstruction(opcodes.F32Sub, sub[float32])
	F32Mul = addInstruction(opcodes.F32Mul, mul[float32])
	F32Eq  = addInstruction(opcodes.F32Eq, eq[float32])
	F32Ne  = addInstruction(opcodes.F32Ne, ne[float32])
	F32Div = addInstruction(opcodes.F32Div, div[float32])
	F32Gt  = addInstruction(opcodes.F32Gt, gt[float32])
	F32Lt  = addInstruction(opcodes.F32Lt, lt[float32])
	F32Abs = addInstruction(opcodes.F32Abs, abs[float32])
	F32Neg = addInstruction(opcodes.F32Neg, neg[float32])
	F32Le  = addInstruction(opcodes.F32Le, le[float32])
	F32Ge  = addInstruction(opcodes.F32Ge, ge[float32])

	F64Add = addInstruction(opcodes.F64Add, add[float64])
	F64Sub = addInstruction(opcodes.F64Sub, sub[float64])
	F64Mul = addInstruction(opcodes.F64Mul, mul[float64])
	F64Eq  = addInstruction(opcodes.F64Eq, eq[float64])
	F64Ne  = addInstruction(opcodes.F64Ne, ne[float64])
	F64Div = addInstruction(opcodes.F64Div, div[float64])
	F64Gt  = addInstruction(opcodes.F64Gt, gt[float64])
	F64Lt  = addInstruction(opcodes.F64Lt, lt[float64])
	F64Abs = addInstruction(opcodes.F64Abs, abs[float64])
	F64Neg = addInstruction(opcodes.F64Neg, neg[float64])
	F64Le  = addInstruction(opcodes.F64Le, le[float64])
	F64Ge  = addInstruction(opcodes.F64Ge, ge[float64])

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

	F32ConvertI32S = addInstruction(opcodes.F32ConvertI32S, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(float32(a))
		return nil
	})

	F32ReinterpretI32 = addInstruction(opcodes.F32ReinterpretI32, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(math.Float32frombits(uint32(a)))
		return nil
	})

	F64ConvertI32S = addInstruction(opcodes.F64ConvertI32S, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(float64(a))
		return nil
	})

	F64Copysign = addInstruction(opcodes.F64Copysign, func(ctx *execution.Context) error {
		sign := stack.Pop[float64](ctx.Stack)
		mag := stack.Pop[float64](ctx.Stack)
		ctx.Stack.Push(math.Copysign(mag, sign))
		return nil
	})

	F64PromoteF32 = addInstruction(opcodes.F64PromoteF32, func(ctx *execution.Context) error {
		a := stack.Pop[float32](ctx.Stack)
		ctx.Stack.Push(float64(a))
		return nil
	})

	F64ReinterpretI64 = addInstruction(opcodes.F64ReinterpretI64, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(math.Float64frombits(uint64(a)))
		return nil
	})

	// i32 sign extension instructions
	I32Extend8S = addInstruction(opcodes.I32Extend8S, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(int32(int8(a)))
		return nil
	})

	I32Extend16S = addInstruction(opcodes.I32Extend16S, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(int32(int16(a)))
		return nil
	})

	// i64 sign extension instructions
	I64Extend8S = addInstruction(opcodes.I64Extend8S, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(int64(int8(a)))
		return nil
	})

	// i32 reinterpret instruction
	I32ReinterpretF32 = addInstruction(opcodes.I32ReinterpretF32, func(ctx *execution.Context) error {
		a := stack.Pop[float32](ctx.Stack)
		ctx.Stack.Push(int32(math.Float32bits(a)))
		return nil
	})

	// i32 truncation instruction
	I32TruncSatF32S = addInstruction(opcodes.I32TruncSatF32S, func(ctx *execution.Context) error {
		a := stack.Pop[float32](ctx.Stack)
		if math.IsNaN(float64(a)) {
			ctx.Stack.Push(int32(0))
		} else if a >= float32(math.MaxInt32) {
			ctx.Stack.Push(int32(math.MaxInt32))
		} else if a <= float32(math.MinInt32) {
			ctx.Stack.Push(int32(math.MinInt32))
		} else {
			ctx.Stack.Push(int32(a))
		}
		return nil
	})

	// i32 wrap instruction
	I32WrapI64 = addInstruction(opcodes.I32WrapI64, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(int32(a))
		return nil
	})

	// i64 sign extension instructions
	I64Extend32S = addInstruction(opcodes.I64Extend32S, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(int64(int32(a)))
		return nil
	})

	I64ExtendI32S = addInstruction(opcodes.I64ExtendI32S, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(int64(a))
		return nil
	})

	I64ExtendI32U = addInstruction(opcodes.I64ExtendI32U, func(ctx *execution.Context) error {
		a := stack.Pop[int32](ctx.Stack)
		ctx.Stack.Push(int64(uint32(a)))
		return nil
	})

	// i64 reinterpret instruction
	I64ReinterpretF64 = addInstruction(opcodes.I64ReinterpretF64, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		ctx.Stack.Push(int64(math.Float64bits(a)))
		return nil
	})

	F64ConvertI64S = addInstruction(opcodes.F64ConvertI64S, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(float64(a))
		return nil
	})

	F64ConvertI64U = addInstruction(opcodes.F64ConvertI64U, func(ctx *execution.Context) error {
		a := stack.Pop[int64](ctx.Stack)
		ctx.Stack.Push(float64(uint64(a)))
		return nil
	})

	F64ConvertI32U = addInstruction(opcodes.F64ConvertI32U, func(ctx *execution.Context) error {
		aVal := stack.Pop[int32](ctx.Stack)
		a := uint32(aVal)
		ctx.Stack.Push(float64(a))
		return nil
	})

	F64Sqrt = addInstruction(opcodes.F64Sqrt, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		ctx.Stack.Push(math.Sqrt(a))
		return nil
	})

	I32TruncSatF64S = addInstruction(opcodes.I32TruncSatF64S, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		if math.IsNaN(a) {
			ctx.Stack.Push(int32(0))
		} else if a >= float64(math.MaxInt32) {
			ctx.Stack.Push(int32(math.MaxInt32))
		} else if a <= float64(math.MinInt32) {
			ctx.Stack.Push(int32(math.MinInt32))
		} else {
			ctx.Stack.Push(int32(a))
		}
		return nil
	})

	I64TruncSatF64S = addInstruction(opcodes.I64TruncSatF64S, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		if math.IsNaN(a) {
			ctx.Stack.Push(int64(0))
		} else if a >= float64(math.MaxInt64) {
			ctx.Stack.Push(int64(math.MaxInt64))
		} else if a <= float64(math.MinInt64) {
			ctx.Stack.Push(int64(math.MinInt64))
		} else {
			ctx.Stack.Push(int64(a))
		}
		return nil
	})

	I64TruncSatF64U = addInstruction(opcodes.I64TruncSatF64U, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		if math.IsNaN(a) || a <= 0 {
			ctx.Stack.Push(int64(0))
		} else if a >= 1.8446744073709552e+19 {
			ctx.Stack.Push(^int64(0))
		} else {
			ctx.Stack.Push(int64(uint64(a)))
		}
		return nil
	})

	I64TruncF64S = addInstruction(opcodes.I64TruncF64S, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		if math.IsNaN(a) {
			return execution.ErrInvalidConversionToInteger
		}
		if math.IsInf(a, 0) || a >= 9.223372036854776e+18 || a < -9.223372036854776e+18 {
			return execution.ErrIntegerOverflow
		}
		ctx.Stack.Push(int64(a))
		return nil
	})

	I64TruncF64U = addInstruction(opcodes.I64TruncF64U, func(ctx *execution.Context) error {
		a := stack.Pop[float64](ctx.Stack)
		if math.IsNaN(a) {
			return execution.ErrInvalidConversionToInteger
		}
		if math.IsInf(a, 0) || a < 0 || a >= 1.8446744073709552e+19 {
			return execution.ErrIntegerOverflow
		}
		ctx.Stack.Push(int64(uint64(a)))
		return nil
	})
)
