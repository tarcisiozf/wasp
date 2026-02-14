package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	MemorySize = addInstruction(opcodes.MemorySize, func(ctx *execution.Context) error {
		index := ctx.Body.Varint()
		ctx.Stack.Push(ctx.Memories[index].Size())
		return nil
	})

	MemoryGrow = addInstruction(opcodes.MemoryGrow, func(ctx *execution.Context) error {
		index := ctx.Body.Varint()
		delta := castInt(ctx.Stack.Pop())
		mem := ctx.Memories[index]
		if mem.Grow(delta) {
			ctx.Stack.Push(1)
		} else {
			ctx.Stack.Push(-1)
		}
		return nil
	})
)
