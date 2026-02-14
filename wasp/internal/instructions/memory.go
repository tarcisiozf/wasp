package instructions

import (
	"encoding/binary"
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

	MemoryLoadI32 = addInstruction(opcodes.MemoryLoadI32, func(ctx *execution.Context) error {
		alignment := ctx.Body.Varint()
		offset := ctx.Body.Varint()
		bytes := ctx.Memories[0].Load(offset, 1<<alignment)
		ctx.Stack.Push(int32(binary.LittleEndian.Uint32(bytes)))
		return nil
	})

	MemoryStoreI32 = addInstruction(opcodes.MemoryStoreI32, func(ctx *execution.Context) error {
		alignment := ctx.Body.Varint()
		offset := ctx.Body.Varint()
		value := castInt(ctx.Stack.Pop())

		bytes := make([]byte, 1<<alignment)
		binary.LittleEndian.PutUint32(bytes, uint32(value))

		ctx.Memories[0].Store(offset, bytes)

		return nil
	})
)
