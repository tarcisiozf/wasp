package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	Nop = addInstruction(opcodes.Nop, func(ctx *execution.Context) error {
		return nil
	})

	Drop = addInstruction(opcodes.Drop, func(ctx *execution.Context) error {
		ctx.Stack.Pop()
		return nil
	})

	If = addInstruction(opcodes.If, func(ctx *execution.Context) error {
		blockType := ctx.Body.Byte()
		ctx.Condition = ctx.Stack.Pop() != 0

		// Push block info for branching
		ctx.BlockStack.Push(execution.BlockInfo{
			Kind:      execution.BlockKindIf,
			StartPos:  ctx.Body.Position(),
			BlockType: blockType,
		})

		if ctx.Condition {
			return nil // execute the if block
		}

		// Skip to else or end
		depth := 1
		for ctx.Body.HasNext() && depth > 0 {
			b := ctx.Body.Byte()
			switch b {
			case opcodes.Block, opcodes.Loop, opcodes.If:
				ctx.Body.Byte() // consume block type
				depth++
			case opcodes.Else:
				if depth == 1 {
					return nil // execute the else block
				}
			case opcodes.End:
				depth--
				if depth == 0 {
					ctx.BlockStack.Pop() // pop our block
					return nil
				}
			default:
				// Skip any immediates for instructions that have them
				skipInstructionImmediates(ctx, b)
			}
		}
		return nil
	})

	Else = addInstruction(opcodes.Else, func(ctx *execution.Context) error {
		// If we reached else, it means the if-block was executed
		// Skip the else block
		depth := 1
		for ctx.Body.HasNext() && depth > 0 {
			b := ctx.Body.Byte()
			switch b {
			case opcodes.Block, opcodes.Loop, opcodes.If:
				ctx.Body.Byte() // consume block type
				depth++
			case opcodes.End:
				depth--
				if depth == 0 {
					ctx.BlockStack.Pop() // pop the if block
					return nil
				}
			default:
				skipInstructionImmediates(ctx, b)
			}
		}
		return nil
	})

	Block = addInstruction(opcodes.Block, func(ctx *execution.Context) error {
		blockType := ctx.Body.Byte()
		ctx.BlockStack.Push(execution.BlockInfo{
			Kind:      execution.BlockKindBlock,
			StartPos:  ctx.Body.Position(),
			BlockType: blockType,
		})
		return nil
	})

	Loop = addInstruction(opcodes.Loop, func(ctx *execution.Context) error {
		blockType := ctx.Body.Byte()
		ctx.BlockStack.Push(execution.BlockInfo{
			Kind:      execution.BlockKindLoop,
			StartPos:  ctx.Body.Position(), // Position after loop header (for br to jump back)
			BlockType: blockType,
		})
		return nil
	})

	Br = addInstruction(opcodes.Br, func(ctx *execution.Context) error {
		labelIdx := ctx.Body.Varint()
		return branchToLabel(ctx, labelIdx)
	})

	BrIf = addInstruction(opcodes.BrIf, func(ctx *execution.Context) error {
		labelIdx := ctx.Body.Varint()
		condition := ctx.Stack.Pop()
		if isNonZero(condition) {
			return branchToLabel(ctx, labelIdx)
		}
		return nil
	})

	Call = addInstruction(opcodes.Call, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		return nil
	})

	End = addInstruction(opcodes.End, func(ctx *execution.Context) error {
		if ctx.BlockStack.Size() == 0 {
			ctx.Done = true
			return nil
		}
		ctx.BlockStack.Pop()
		return nil
	})
)

// branchToLabel implements the branch operation
// For blocks and if: branch to end of block
// For loops: branch to start of loop (re-execute)
func branchToLabel(ctx *execution.Context, labelIdx int) error {
	// Get the target block (labelIdx is relative depth, 0 = innermost)
	stackSize := ctx.BlockStack.Size()
	targetIdx := stackSize - 1 - labelIdx
	targetBlock := ctx.BlockStack.At(targetIdx)

	// Pop all blocks up to and including the target
	for i := 0; i <= labelIdx; i++ {
		ctx.BlockStack.Pop()
	}

	if targetBlock.Kind == execution.BlockKindLoop {
		// For loops, jump back to the start and re-push the block
		ctx.Body.Seek(targetBlock.StartPos)
		ctx.BlockStack.Push(targetBlock)
		return nil
	}

	// For blocks and if, skip to the end
	// depth accounts for all the blocks we're skipping out of (labelIdx + 1)
	depth := labelIdx + 1
	for ctx.Body.HasNext() && depth > 0 {
		b := ctx.Body.Byte()
		switch b {
		case opcodes.Block, opcodes.Loop, opcodes.If:
			ctx.Body.Byte() // consume block type
			depth++
		case opcodes.End:
			depth--
		default:
			skipInstructionImmediates(ctx, b)
		}
	}
	return nil
}

// isNonZero checks if a value is non-zero, handling different integer types
func isNonZero(v any) bool {
	switch val := v.(type) {
	case int:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case uint32:
		return val != 0
	case uint64:
		return val != 0
	default:
		return v != 0
	}
}

// skipInstructionImmediates skips the immediate values for instructions
// that have them, when we're skipping over code (e.g., in if/else/br)
func skipInstructionImmediates(ctx *execution.Context, opcode byte) {
	switch opcode {
	// Control instructions with label index
	case opcodes.Br, opcodes.BrIf:
		ctx.Body.Varint() // label index
	case opcodes.Call:
		ctx.Body.Varint() // function index

	// Variable instructions
	case opcodes.LocalGet, opcodes.LocalSet, opcodes.LocalTee:
		ctx.Body.Varint() // local index
	case opcodes.GlobalGet, opcodes.GlobalSet:
		ctx.Body.Varint() // global index

	// Const instructions
	case opcodes.I32Const:
		ctx.Body.Varint() // i32 value
	case opcodes.I64Const:
		ctx.Body.Varint() // i64 value
	case opcodes.F32Const:
		ctx.Body.Bytes(4) // f32 value (4 bytes)
	case opcodes.F64Const:
		ctx.Body.Bytes(8) // f64 value (8 bytes)

	// Memory instructions
	case opcodes.MemoryLoadI32, opcodes.MemoryStoreI32:
		ctx.Body.Varint() // align
		ctx.Body.Varint() // offset
	case opcodes.MemorySize, opcodes.MemoryGrow:
		ctx.Body.Byte() // memory index (always 0x00)

		// Most other instructions have no immediates
	}
}
