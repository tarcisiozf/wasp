package instructions_test

import (
	"testing"
	"wasp/wasp/internal/binary/leb"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/memory"

	"github.com/stretchr/testify/assert"
)

// createTestContextWithBlocks creates a test context with BlockStack initialized
func createTestContextWithBlocks(body ...[]byte) *execution.Context {
	ctx := createTestContext(body...)
	ctx.BlockStack = memory.NewStackWithCapacity[execution.BlockFrame](16)
	ctx.Blocks = make(map[int]fnblock.Target)
	return ctx
}

func TestUnreachable(t *testing.T) {
	ctx := createTestContext()

	err := instructions.Unreachable.Handler(ctx)
	assert.ErrorIs(t, err, execution.ErrUnreachable)
}

func TestNop(t *testing.T) {
	ctx := createTestContext()

	err := instructions.Nop.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())
}

func TestDrop(t *testing.T) {
	ctx := createTestContext()
	ctx.Stack.Push(int32(42))
	ctx.Stack.Push(int32(100))

	err := instructions.Drop.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(42), ctx.Stack.Top())
}

func TestReturnUnit(t *testing.T) {
	ctx := createTestContext()
	ctx.Done = false

	err := instructions.Return.Handler(ctx)
	assert.NoError(t, err)
	assert.True(t, ctx.Done)
}

func TestSelectUnit(t *testing.T) {
	t.Run("condition true selects first value", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10)) // val1
		ctx.Stack.Push(int32(20)) // val2
		ctx.Stack.Push(int32(1))  // condition (non-zero = true)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(10), ctx.Stack.Top())
	})

	t.Run("condition false selects second value", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10)) // val1
		ctx.Stack.Push(int32(20)) // val2
		ctx.Stack.Push(int32(0))  // condition (zero = false)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(20), ctx.Stack.Top())
	})

	t.Run("condition with int64", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))  // val1
		ctx.Stack.Push(int32(20))  // val2
		ctx.Stack.Push(int64(100)) // condition (non-zero int64)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(10), ctx.Stack.Top())
	})
}

func TestCallUnit(t *testing.T) {
	functionIndex := 5
	ctx := createTestContext(
		leb.Encode(functionIndex),
	)

	err := instructions.Call.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, functionIndex, ctx.FunctionCallRequest)
}

func TestReturnCallUnit(t *testing.T) {
	functionIndex := 7
	ctx := createTestContext(
		leb.Encode(functionIndex),
	)

	err := instructions.ReturnCall.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, functionIndex, ctx.FunctionCallRequest)
	assert.True(t, ctx.TailCall)
}

func TestEnd(t *testing.T) {
	t.Run("with empty block stack sets done", func(t *testing.T) {
		ctx := createTestContextWithBlocks()

		err := instructions.End.Handler(ctx)
		assert.NoError(t, err)
		assert.True(t, ctx.Done)
	})

	t.Run("with blocks pops block stack", func(t *testing.T) {
		ctx := createTestContextWithBlocks()
		ctx.BlockStack.Push(execution.BlockFrame{StartPos: 10})

		err := instructions.End.Handler(ctx)
		assert.NoError(t, err)
		assert.False(t, ctx.Done)
		assert.Equal(t, 0, ctx.BlockStack.Size())
	})
}

func TestBlock(t *testing.T) {
	// Block instruction consumes 1 byte for block type, then pushes frame
	ctx := createTestContextWithBlocks(
		[]byte{0x40}, // void block type
	)

	err := instructions.Block.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.BlockStack.Size())
}

func TestLoop(t *testing.T) {
	// Loop instruction consumes 1 byte for block type, then pushes frame
	ctx := createTestContextWithBlocks(
		[]byte{0x40}, // void block type
	)

	err := instructions.Loop.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.BlockStack.Size())
}

func TestCallIndirectUnit(t *testing.T) {
	t.Run("valid call", func(t *testing.T) {
		signatureIndex := 0
		tableIndex := 0
		ctx := createTestContext(
			leb.Encode(signatureIndex),
			leb.Encode(tableIndex),
		)

		// Setup table with function reference
		table := &memory.Table{
			Elements: []int{2}, // function index 2 at element 0
		}
		ctx.Tables = []*memory.Table{table}

		// Setup function signatures - both pointing to same signature
		ctx.FuncSignatures = []fnsig.Signature{
			{Params: []byte{}, Results: []byte{}},
			{Params: []byte{}, Results: []byte{}},
			{Params: []byte{}, Results: []byte{}}, // function 2
		}

		// Push element index
		ctx.Stack.Push(int32(0))

		err := instructions.CallIndirect.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, ctx.FunctionCallRequest)
	})

	t.Run("invalid table index", func(t *testing.T) {
		signatureIndex := 0
		tableIndex := 5 // invalid
		ctx := createTestContext(
			leb.Encode(signatureIndex),
			leb.Encode(tableIndex),
		)
		ctx.Tables = []*memory.Table{}
		ctx.Stack.Push(int32(0))

		err := instructions.CallIndirect.Handler(ctx)
		assert.ErrorIs(t, err, execution.ErrInvalidTableIndex)
	})

	t.Run("undefined element", func(t *testing.T) {
		signatureIndex := 0
		tableIndex := 0
		ctx := createTestContext(
			leb.Encode(signatureIndex),
			leb.Encode(tableIndex),
		)

		table := &memory.Table{
			Elements: []int{1}, // only 1 element
		}
		ctx.Tables = []*memory.Table{table}
		ctx.Stack.Push(int32(5)) // element index out of bounds

		err := instructions.CallIndirect.Handler(ctx)
		assert.ErrorIs(t, err, execution.ErrUndefinedElement)
	})

	t.Run("uninitialized element", func(t *testing.T) {
		signatureIndex := 0
		tableIndex := 0
		ctx := createTestContext(
			leb.Encode(signatureIndex),
			leb.Encode(tableIndex),
		)

		table := &memory.Table{
			Elements: []int{-1}, // uninitialized
		}
		ctx.Tables = []*memory.Table{table}
		ctx.FuncSignatures = []fnsig.Signature{{}}
		ctx.Stack.Push(int32(0))

		err := instructions.CallIndirect.Handler(ctx)
		assert.ErrorIs(t, err, execution.ErrUninitializedElement)
	})

	t.Run("signature mismatch", func(t *testing.T) {
		signatureIndex := 0
		tableIndex := 0
		ctx := createTestContext(
			leb.Encode(signatureIndex),
			leb.Encode(tableIndex),
		)

		table := &memory.Table{
			Elements: []int{1}, // function index 1
		}
		ctx.Tables = []*memory.Table{table}

		// Different signatures
		ctx.FuncSignatures = []fnsig.Signature{
			{Params: []byte{0x7F}, Results: []byte{}}, // sig 0: (i32) -> ()
			{Params: []byte{}, Results: []byte{0x7F}}, // sig 1: () -> (i32) - different!
		}
		ctx.Stack.Push(int32(0))

		err := instructions.CallIndirect.Handler(ctx)
		assert.ErrorIs(t, err, execution.ErrIndirectCallTypeMismatch)
	})
}

func TestBrIfUnit(t *testing.T) {
	t.Run("condition false - no branch", func(t *testing.T) {
		ctx := createTestContextWithBlocks(
			leb.Encode(0), // label index
		)
		ctx.BlockStack.Push(execution.BlockFrame{StartPos: 0})
		ctx.Blocks[0] = fnblock.Target{Kind: fnblock.KindBlock, EndPos: 100}

		ctx.Stack.Push(int32(0)) // false condition

		err := instructions.BrIf.Handler(ctx)
		assert.NoError(t, err)
		// Should not have branched - block stack unchanged
		assert.Equal(t, 1, ctx.BlockStack.Size())
	})

	t.Run("condition true - branch taken", func(t *testing.T) {
		ctx := createTestContextWithBlocks(
			leb.Encode(0), // label index
		)
		ctx.BlockStack.Push(execution.BlockFrame{StartPos: 0})
		ctx.Blocks[0] = fnblock.Target{Kind: fnblock.KindBlock, EndPos: 100}

		ctx.Stack.Push(int32(1)) // true condition

		err := instructions.BrIf.Handler(ctx)
		assert.NoError(t, err)
		// Should have branched - block popped
		assert.Equal(t, 0, ctx.BlockStack.Size())
		assert.Equal(t, 100, ctx.Body.Position())
	})
}

func TestBr(t *testing.T) {
	ctx := createTestContextWithBlocks(
		leb.Encode(0), // label index (innermost block)
	)
	ctx.BlockStack.Push(execution.BlockFrame{StartPos: 5})
	ctx.Blocks[5] = fnblock.Target{Kind: fnblock.KindBlock, EndPos: 50}

	err := instructions.Br.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.BlockStack.Size())
	assert.Equal(t, 50, ctx.Body.Position())
}

func TestIf(t *testing.T) {
	t.Run("condition true - execute if block", func(t *testing.T) {
		ctx := createTestContextWithBlocks(
			[]byte{0x40}, // block type (void)
		)
		ctx.Stack.Push(int32(1)) // true condition

		err := instructions.If.Handler(ctx)
		assert.NoError(t, err)
		assert.True(t, ctx.Condition)
		assert.Equal(t, 1, ctx.BlockStack.Size())
	})

	t.Run("condition false with else - jump to else", func(t *testing.T) {
		ctx := createTestContextWithBlocks(
			[]byte{0x40}, // block type (void)
		)
		ctx.Stack.Push(int32(0)) // false condition

		// Setup precomputed target with else position
		// The startPos will be 1 (after consuming block type byte)
		ctx.Blocks[1] = fnblock.Target{
			Kind:    fnblock.KindIf,
			ElsePos: 50,
			EndPos:  100,
		}

		err := instructions.If.Handler(ctx)
		assert.NoError(t, err)
		assert.False(t, ctx.Condition)
		assert.Equal(t, 50, ctx.Body.Position())
		assert.Equal(t, 1, ctx.BlockStack.Size()) // block frame still pushed for else
	})

	t.Run("condition false without else - jump to end", func(t *testing.T) {
		ctx := createTestContextWithBlocks(
			[]byte{0x40}, // block type (void)
		)
		ctx.Stack.Push(int32(0)) // false condition

		// Setup precomputed target without else
		ctx.Blocks[1] = fnblock.Target{
			Kind:    fnblock.KindIf,
			ElsePos: 0, // no else
			EndPos:  100,
		}

		err := instructions.If.Handler(ctx)
		assert.NoError(t, err)
		assert.False(t, ctx.Condition)
		assert.Equal(t, 100, ctx.Body.Position())
		assert.Equal(t, 0, ctx.BlockStack.Size()) // block frame popped
	})
}

func TestElse(t *testing.T) {
	ctx := createTestContextWithBlocks()

	// Push a block frame that would be created by If
	ctx.BlockStack.Push(execution.BlockFrame{StartPos: 10})
	ctx.Blocks[10] = fnblock.Target{
		Kind:   fnblock.KindIf,
		EndPos: 100,
	}

	err := instructions.Else.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 100, ctx.Body.Position())
	assert.Equal(t, 0, ctx.BlockStack.Size())
}

func TestSelectWithFloats(t *testing.T) {
	t.Run("condition with float32 non-zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))    // val1
		ctx.Stack.Push(int32(20))    // val2
		ctx.Stack.Push(float32(1.5)) // condition (non-zero float32)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(10), ctx.Stack.Top())
	})

	t.Run("condition with float32 zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))    // val1
		ctx.Stack.Push(int32(20))    // val2
		ctx.Stack.Push(float32(0.0)) // condition (zero float32)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(20), ctx.Stack.Top())
	})

	t.Run("condition with float64 non-zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))    // val1
		ctx.Stack.Push(int32(20))    // val2
		ctx.Stack.Push(float64(2.5)) // condition (non-zero float64)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(10), ctx.Stack.Top())
	})

	t.Run("condition with float64 zero", func(t *testing.T) {
		ctx := createTestContext()
		ctx.Stack.Push(int32(10))    // val1
		ctx.Stack.Push(int32(20))    // val2
		ctx.Stack.Push(float64(0.0)) // condition (zero float64)

		err := instructions.Select.Handler(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, ctx.Stack.Size())
		assert.Equal(t, int32(20), ctx.Stack.Top())
	})
}

func TestBrLoop(t *testing.T) {
	// Test that branching to a loop jumps back to start
	ctx := createTestContextWithBlocks(
		leb.Encode(0), // label index (innermost = loop)
	)

	loopStartPos := 10
	ctx.BlockStack.Push(execution.BlockFrame{StartPos: loopStartPos})
	ctx.Blocks[loopStartPos] = fnblock.Target{
		Kind:     fnblock.KindLoop,
		StartPos: loopStartPos,
		EndPos:   50,
	}

	err := instructions.Br.Handler(ctx)
	assert.NoError(t, err)
	// For loops, branch jumps back to start and re-pushes frame
	assert.Equal(t, 1, ctx.BlockStack.Size())
	assert.Equal(t, loopStartPos, ctx.Body.Position())
}
