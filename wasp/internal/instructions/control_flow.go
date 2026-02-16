package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/opcodes"
)

var (
	Unreachable = addInstruction(opcodes.Unreachable, func(ctx *execution.Context) error {
		return execution.ErrUnreachable
	})

	Nop = addInstruction(opcodes.Nop, func(ctx *execution.Context) error {
		return nil
	})

	Drop = addInstruction(opcodes.Drop, func(ctx *execution.Context) error {
		ctx.Stack.Pop()
		return nil
	})

	If = addInstruction(opcodes.If, func(ctx *execution.Context) error {
		_ = ctx.Body.Byte() // consume block type
		startPos := ctx.Body.Position()
		ctx.Condition = ctx.Stack.Pop() != 0

		// Push block frame for branching
		ctx.BlockStack.Push(execution.BlockFrame{
			StartPos: startPos,
		})

		if ctx.Condition {
			return nil // execute the if block
		}

		// Use precomputed targets to jump
		target := ctx.Blocks[startPos]
		if target.ElsePos != 0 {
			ctx.Body.Seek(target.ElsePos) // jump to else
		} else {
			ctx.Body.Seek(target.EndPos) // jump past end
			ctx.BlockStack.Pop()
		}
		return nil
	})

	Else = addInstruction(opcodes.Else, func(ctx *execution.Context) error {
		// If we reached else, it means the if-block was executed
		// Use precomputed target to skip the else block
		frame := ctx.BlockStack.Peek()
		target := ctx.Blocks[frame.StartPos]
		ctx.Body.Seek(target.EndPos)
		ctx.BlockStack.Pop()
		return nil
	})

	Block = addInstruction(opcodes.Block, func(ctx *execution.Context) error {
		_ = ctx.Body.Byte() // consume block type
		ctx.BlockStack.Push(execution.BlockFrame{
			StartPos: ctx.Body.Position(),
		})
		return nil
	})

	Loop = addInstruction(opcodes.Loop, func(ctx *execution.Context) error {
		_ = ctx.Body.Byte() // consume block type
		ctx.BlockStack.Push(execution.BlockFrame{
			StartPos: ctx.Body.Position(), // Position after loop header (for br to jump back)
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

	BrTable = addInstruction(opcodes.BrTable, func(ctx *execution.Context) error {
		// Read the number of targets
		numTargets := ctx.Body.Varint()

		// Read all target labels
		targets := make([]int, numTargets)
		for i := 0; i < numTargets; i++ {
			targets[i] = ctx.Body.Varint()
		}

		// Read the default label
		defaultLabel := ctx.Body.Varint()

		// Pop the index from the stack
		index := castInt(ctx.Stack.Pop())

		// Select the target label
		var labelIdx int
		if index >= 0 && index < numTargets {
			labelIdx = targets[index]
		} else {
			labelIdx = defaultLabel
		}

		return branchToLabel(ctx, labelIdx)
	})

	Return = addInstruction(opcodes.Return, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})

	ReturnCall = addInstruction(opcodes.ReturnCall, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		ctx.TailCall = true
		return nil
	})

	Call = addInstruction(opcodes.Call, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		return nil
	})

	CallIndirect = addInstruction(opcodes.CallIndirect, func(ctx *execution.Context) error {
		signatureIndex := ctx.Body.Varint()
		tableIndex := ctx.Body.Varint()

		// Pop the element index from the stack
		elementIndex := castInt(ctx.Stack.Pop())

		// Validate table index
		if tableIndex < 0 || tableIndex >= len(ctx.Tables) {
			return execution.ErrInvalidTableIndex
		}
		table := ctx.Tables[tableIndex]

		// Validate element index bounds
		if elementIndex < 0 || elementIndex >= len(table.Elements) {
			return execution.ErrUndefinedElement
		}

		// Get the function index from the table
		funcIndex := table.Elements[elementIndex]

		// Check if the element is initialized (-1 means uninitialized)
		if funcIndex < 0 {
			return execution.ErrUninitializedElement
		}

		if signatureIndex < 0 || signatureIndex >= len(ctx.FuncSignatures) {
			return execution.ErrInvalidSignatureIndex
		}
		expectedSig := ctx.FuncSignatures[signatureIndex]

		if funcIndex < 0 || funcIndex >= len(ctx.FuncSignatures) {
			return execution.ErrInvalidFunctionIndex
		}
		funcSig := ctx.FuncSignatures[funcIndex]

		if !expectedSig.Equals(funcSig) {
			return execution.ErrIndirectCallTypeMismatch
		}

		ctx.FunctionCallRequest = funcIndex

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

	Select = addInstruction(opcodes.Select, func(ctx *execution.Context) error {
		condition := ctx.Stack.Pop()
		val2 := ctx.Stack.Pop()
		val1 := ctx.Stack.Pop()
		if isNonZero(condition) {
			ctx.Stack.Push(val1)
		} else {
			ctx.Stack.Push(val2)
		}
		return nil
	})
)

// branchToLabel implements the branch operation using precomputed targets
// For fnblock and if: branch to end of block
// For loops: branch to start of loop (re-execute)
func branchToLabel(ctx *execution.Context, labelIdx int) error {
	// Get the target block (labelIdx is relative depth, 0 = innermost)
	stackSize := ctx.BlockStack.Size()
	targetIdx := stackSize - 1 - labelIdx
	targetFrame := ctx.BlockStack.At(targetIdx)
	target := ctx.Blocks[targetFrame.StartPos]

	// Pop all fnblock up to and including the target
	for i := 0; i <= labelIdx; i++ {
		ctx.BlockStack.Pop()
	}

	if target.Kind == fnblock.KindLoop {
		// For loops, jump back to the start and re-push the frame
		ctx.Body.Seek(target.StartPos)
		ctx.BlockStack.Push(targetFrame)
		return nil
	}

	// For fnblock and if, jump directly to the end using precomputed position
	ctx.Body.Seek(target.EndPos)
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
