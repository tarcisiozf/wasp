package instructions_test

import (
	"testing"
	"wasp/wasp/internal/binary/leb"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/memory"

	"github.com/stretchr/testify/assert"
)

func TestLocalGet(t *testing.T) {
	localIndex := 0
	ctx := createTestContext(
		leb.Encode(localIndex),
	)

	n := 42
	ctx.Locals.Push(n)

	err := instructions.LocalGet.Handler(ctx)
	assert.NoError(t, err)

	result := ctx.Stack.Top()
	assert.Equal(t, n, result)
	assert.Equal(t, 1, ctx.Stack.Size())
}

func TestLocalSet(t *testing.T) {
	localIndex := 0
	ctx := createTestContext(
		leb.Encode(localIndex),
	)

	n := 42
	ctx.Locals.Push(-1)
	ctx.Stack.Push(n)

	err := instructions.LocalSet.Handler(ctx)
	assert.NoError(t, err)
	assert.Zero(t, ctx.Stack.Size())
	assert.Equal(t, 1, ctx.Locals.Size())
	assert.Equal(t, n, ctx.Locals.At(localIndex))
}

func TestLocalTee(t *testing.T) {
	localIndex := 0
	ctx := createTestContext(
		leb.Encode(localIndex),
	)

	n := 42
	ctx.Locals.Push(-1)
	ctx.Stack.Push(n)

	err := instructions.LocalTee.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, n, ctx.Stack.Top())
	assert.Equal(t, 1, ctx.Locals.Size())
	assert.Equal(t, n, ctx.Locals.At(localIndex))
}

func TestGlobalGet(t *testing.T) {
	globalIndex := 0
	ctx := createTestContext(
		leb.Encode(globalIndex),
	)

	n := 42
	ctx.Globals.Push(n, false)

	err := instructions.GlobalGet.Handler(ctx)
	assert.NoError(t, err)

	result := ctx.Stack.Top()
	assert.Equal(t, n, result)
	assert.Equal(t, 1, ctx.Stack.Size())
}

func TestGlobalSet(t *testing.T) {
	const globalIndex = 0
	const n = 42

	t.Run("mutable", func(t *testing.T) {
		ctx := createTestContext(
			leb.Encode(globalIndex),
		)

		ctx.Globals.Push(0, true)
		ctx.Stack.Push(n)

		err := instructions.GlobalSet.Handler(ctx)
		assert.NoError(t, err)

		result, err := ctx.Globals.Get(globalIndex)
		assert.NoError(t, err)
		assert.Equal(t, n, result)
		assert.Zero(t, ctx.Stack.Size())
	})

	t.Run("try to set immutable", func(t *testing.T) {
		ctx := createTestContext(
			leb.Encode(globalIndex),
		)

		ctx.Globals.Push(0, false)
		ctx.Stack.Push(n)

		err := instructions.GlobalSet.Handler(ctx)
		assert.ErrorIs(t, err, memory.ErrCannotSetImmutableGlobal)

		result, err := ctx.Globals.Get(globalIndex)
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
		assert.Zero(t, ctx.Stack.Size())
	})
}
