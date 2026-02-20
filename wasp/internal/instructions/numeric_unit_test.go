package instructions_test

import (
	"math"
	"testing"
	"wasp/wasp/internal/instructions"

	"github.com/stretchr/testify/assert"
)

func TestI32Add(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int32(10))
	ctx.Stack.Push(int32(20))

	err := instructions.I32Add.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(30), ctx.Stack.Top())
}

func TestI32Sub(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int32(50))
	ctx.Stack.Push(int32(20))

	err := instructions.I32Sub.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(30), ctx.Stack.Top())
}

func TestI32Mul(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int32(6))
	ctx.Stack.Push(int32(7))

	err := instructions.I32Mul.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(42), ctx.Stack.Top())
}

func TestI32Div(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int32(100))
	ctx.Stack.Push(int32(10))

	err := instructions.I32Div.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(10), ctx.Stack.Top())
}

func TestI32Eq(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(42))
		ctx.Stack.Push(int32(42))

		err := instructions.I32Eq.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("not equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(42))
		ctx.Stack.Push(int32(43))

		err := instructions.I32Eq.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32Eqz(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(0))

		err := instructions.I32Eqz.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("non-zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(42))

		err := instructions.I32Eqz.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32Ne(t *testing.T) {
	t.Run("not equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(42))
		ctx.Stack.Push(int32(43))

		err := instructions.I32Ne.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(42))
		ctx.Stack.Push(int32(42))

		err := instructions.I32Ne.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32And(t *testing.T) {
	ctx := createTestContext()
	ctx.Stack.Push(int32(0b1010))
	ctx.Stack.Push(int32(0b1100))

	err := instructions.I32And.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(0b1000), ctx.Stack.Top())
}

func TestI32Or(t *testing.T) {
	ctx := createTestContext()
	ctx.Stack.Push(int32(0b1010))
	ctx.Stack.Push(int32(0b1100))

	err := instructions.I32Or.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(0b1110), ctx.Stack.Top())
}

func TestI32Xor(t *testing.T) {
	ctx := createTestContext()
	ctx.Stack.Push(int32(0b1010))
	ctx.Stack.Push(int32(0b1100))

	err := instructions.I32Xor.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(0b0110), ctx.Stack.Top())
}

func TestI32LtS(t *testing.T) {
	t.Run("less than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(20))

		err := instructions.I32LtS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("not less than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(20))
		ctx.Stack.Push(int32(10))

		err := instructions.I32LtS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})

	t.Run("negative numbers", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(-10))
		ctx.Stack.Push(int32(5))

		err := instructions.I32LtS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})
}

func TestI32GtS(t *testing.T) {
	t.Run("greater than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(20))
		ctx.Stack.Push(int32(10))

		err := instructions.I32GtS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("not greater than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(20))

		err := instructions.I32GtS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32LeS(t *testing.T) {
	t.Run("less than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(20))

		err := instructions.I32LeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(10))

		err := instructions.I32LeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("greater than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(20))
		ctx.Stack.Push(int32(10))

		err := instructions.I32LeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32GeS(t *testing.T) {
	t.Run("greater than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(20))
		ctx.Stack.Push(int32(10))

		err := instructions.I32GeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(10))

		err := instructions.I32GeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("less than", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(20))

		err := instructions.I32GeS.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI32LtU(t *testing.T) {
	t.Run("unsigned comparison", func(t *testing.T) {
		ctx := createTestContext()
		// -1 as int32 is max uint32, so it should be greater than 10 in unsigned comparison
		ctx.Stack.Push(int32(10))
		ctx.Stack.Push(int32(-1))

		err := instructions.I32LtU.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top()) // 10 < 0xFFFFFFFF (unsigned)
	})
}

func TestI64Add(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int64(10))
	ctx.Stack.Push(int64(20))

	err := instructions.I64Add.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int64(30), ctx.Stack.Top())
}

func TestI64Sub(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int64(50))
	ctx.Stack.Push(int64(20))

	err := instructions.I64Sub.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int64(30), ctx.Stack.Top())
}

func TestI64Mul(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int64(6))
	ctx.Stack.Push(int64(7))

	err := instructions.I64Mul.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int64(42), ctx.Stack.Top())
}

func TestI64Div(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(int64(100))
	ctx.Stack.Push(int64(10))

	err := instructions.I64Div.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int64(10), ctx.Stack.Top())
}

func TestI64Eq(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int64(42))
		ctx.Stack.Push(int64(42))

		err := instructions.I64Eq.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("not equal", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int64(42))
		ctx.Stack.Push(int64(43))

		err := instructions.I64Eq.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestI64Eqz(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int64(0))

		err := instructions.I64Eqz.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 1, ctx.Stack.Top())
	})

	t.Run("non-zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int64(42))

		err := instructions.I64Eqz.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, 0, ctx.Stack.Top())
	})
}

func TestF32Add(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(float32(10.5))
	ctx.Stack.Push(float32(20.5))

	err := instructions.F32Add.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, float32(31.0), ctx.Stack.Top())
}

func TestF32Sub(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(float32(50.5))
	ctx.Stack.Push(float32(20.5))

	err := instructions.F32Sub.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, float32(30.0), ctx.Stack.Top())
}

func TestF64Mul(t *testing.T) {
	ctx := createTestContext()

	ctx.Stack.Push(float64(6.0))
	ctx.Stack.Push(float64(7.0))

	err := instructions.F64Mul.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, float64(42.0), ctx.Stack.Top())
}

func TestF32Abs(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(float32(42.5))

		err := instructions.F32Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, float32(42.5), ctx.Stack.Top())
	})

	t.Run("negative", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(float32(-42.5))

		err := instructions.F32Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, float32(42.5), ctx.Stack.Top())
	})

	t.Run("negative zero", func(t *testing.T) {
		ctx := createTestContext()
		negZero := float32(math.Copysign(0, -1))
		ctx.Stack.Push(negZero)

		err := instructions.F32Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		// Result should be positive zero
		result := ctx.Stack.Top().(float32)
		assert.False(t, math.Signbit(float64(result)))
	})
}

func TestF64Abs(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(float64(42.5))

		err := instructions.F64Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, float64(42.5), ctx.Stack.Top())
	})

	t.Run("negative", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(float64(-42.5))

		err := instructions.F64Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, float64(42.5), ctx.Stack.Top())
	})

	t.Run("negative zero", func(t *testing.T) {
		ctx := createTestContext()
		negZero := math.Copysign(0, -1)
		ctx.Stack.Push(negZero)

		err := instructions.F64Abs.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		// Result should be positive zero
		result := ctx.Stack.Top().(float64)
		assert.False(t, math.Signbit(result))
	})
}
