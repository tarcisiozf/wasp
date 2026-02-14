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
		ctx.BlockType = ctx.Body.Byte()
		ctx.Condition = ctx.Stack.Pop() != 0
		if ctx.Condition {
			return nil // execute the if block
		}

		// Skip the block
	loop:
		for ctx.Body.HasNext() {
			b := ctx.Body.Peek()
			switch b {
			case opcodes.Else:
				break loop
			case opcodes.End:
				ctx.Body.Next() // consume the end opcode
				break loop
			}
			ctx.Body.Next()
		}
		return nil
	})

	Else = addInstruction(opcodes.Else, func(ctx *execution.Context) error {
		if !ctx.Condition {
			// Skip the else block
			for ctx.Body.HasNext() {
				if ctx.Body.Byte() == opcodes.End {
					break
				}
			}
		}
		return nil
	})

	Block = addInstruction(opcodes.Block, func(ctx *execution.Context) error {
		ctx.BlockType = ctx.Body.Byte()
		ctx.Depth++
		return nil
	})

	Br = addInstruction(opcodes.Br, func(ctx *execution.Context) error {
		breakDepth := ctx.Body.Varint()
		// Skip instructions until we find the matching end of the block
		depth := ctx.Depth
		for depth != breakDepth && ctx.Body.HasNext() {
			b := ctx.Body.Byte()
			if b == opcodes.Block || b == opcodes.If {
				depth++
			} else if b == opcodes.End {
				depth--
			}
		}
		return nil
	})

	Call = addInstruction(opcodes.Call, func(ctx *execution.Context) error {
		functionIndex := ctx.Body.Varint()
		ctx.FunctionCallRequest = functionIndex
		return nil
	})

	End = addInstruction(opcodes.End, func(ctx *execution.Context) error {
		ctx.Done = true
		return nil
	})
)
