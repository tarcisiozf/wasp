package instructions

import (
	"encoding/binary"
	"fmt"
	"math"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

func castInt64(item any) int64 {
	switch v := item.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	}
	panic(fmt.Sprintf("expected int64, got %T", item))
}

var (
	MemorySize = addInstruction(opcodes.MemorySize, func(ctx *execution.Context) error {
		index := ctx.Body.Varint()
		ctx.Stack.Push(ctx.Memories[index].NumPages())
		return nil
	})

	MemoryGrow = addInstruction(opcodes.MemoryGrow, func(ctx *execution.Context) error {
		index := ctx.Body.Varint()
		delta := castInt(ctx.Stack.Pop())
		mem := ctx.Memories[index]
		prevPages := mem.NumPages()
		if mem.Grow(delta) {
			ctx.Stack.Push(prevPages)
		} else {
			ctx.Stack.Push(-1)
		}
		return nil
	})

	MemoryCopy = addInstruction(opcodes.MemoryCopy, func(ctx *execution.Context) error {
		dstIndex := ctx.Body.Varint()
		srcIndex := ctx.Body.Varint()

		size := castInt(ctx.Stack.Pop())
		srcOffset := castInt(ctx.Stack.Pop())
		dstOffset := castInt(ctx.Stack.Pop())

		src := ctx.Memories[srcIndex]
		dst := ctx.Memories[dstIndex]

		dst.Store(dstOffset, src.Load(srcOffset, size))

		return nil
	})

	MemoryFill = addInstruction(opcodes.MemoryFill, func(ctx *execution.Context) error {
		index := ctx.Body.Varint()
		size := castInt(ctx.Stack.Pop())
		value := castInt(ctx.Stack.Pop())
		offset := castInt(ctx.Stack.Pop())

		if value < 0 || value > 255 {
			return fmt.Errorf("memory.fill value must be between 0 and 255, got %d", value)
		}

		chunk := make([]byte, size)
		for i := range chunk {
			chunk[i] = byte(value)
		}

		mem := ctx.Memories[index]
		mem.Store(offset, chunk)

		return nil
	})

	// i32.load - load 4 bytes as i32
	I32Load = addInstruction(opcodes.I32Load, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment (unused, just for validation hints)
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 4)
		ctx.Stack.Push(int32(binary.LittleEndian.Uint32(bytes)))
		return nil
	})

	// i64.load - load 8 bytes as i64
	I64Load = addInstruction(opcodes.I64Load, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 8)
		ctx.Stack.Push(int64(binary.LittleEndian.Uint64(bytes)))
		return nil
	})

	// f32.load - load 4 bytes as f32
	F32Load = addInstruction(opcodes.F32Load, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 4)
		bits := binary.LittleEndian.Uint32(bytes)
		ctx.Stack.Push(math.Float32frombits(bits))
		return nil
	})

	// f64.load - load 8 bytes as f64
	F64Load = addInstruction(opcodes.F64Load, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 8)
		bits := binary.LittleEndian.Uint64(bytes)
		ctx.Stack.Push(math.Float64frombits(bits))
		return nil
	})

	// i32.load8_s - load 1 byte, sign-extend to i32
	I32Load8S = addInstruction(opcodes.I32Load8S, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 1)
		ctx.Stack.Push(int32(int8(bytes[0])))
		return nil
	})

	// i32.load8_u - load 1 byte, zero-extend to i32
	I32Load8U = addInstruction(opcodes.I32Load8U, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 1)
		ctx.Stack.Push(int32(bytes[0]))
		return nil
	})

	// i32.load16_s - load 2 bytes, sign-extend to i32
	I32Load16S = addInstruction(opcodes.I32Load16S, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 2)
		ctx.Stack.Push(int32(int16(binary.LittleEndian.Uint16(bytes))))
		return nil
	})

	// i32.load16_u - load 2 bytes, zero-extend to i32
	I32Load16U = addInstruction(opcodes.I32Load16U, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 2)
		ctx.Stack.Push(int32(binary.LittleEndian.Uint16(bytes)))
		return nil
	})

	// i64.load8_s - load 1 byte, sign-extend to i64
	I64Load8S = addInstruction(opcodes.I64Load8S, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 1)
		ctx.Stack.Push(int64(int8(bytes[0])))
		return nil
	})

	// i64.load8_u - load 1 byte, zero-extend to i64
	I64Load8U = addInstruction(opcodes.I64Load8U, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 1)
		ctx.Stack.Push(int64(bytes[0]))
		return nil
	})

	// i64.load16_s - load 2 bytes, sign-extend to i64
	I64Load16S = addInstruction(opcodes.I64Load16S, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 2)
		ctx.Stack.Push(int64(int16(binary.LittleEndian.Uint16(bytes))))
		return nil
	})

	// i64.load16_u - load 2 bytes, zero-extend to i64
	I64Load16U = addInstruction(opcodes.I64Load16U, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 2)
		ctx.Stack.Push(int64(binary.LittleEndian.Uint16(bytes)))
		return nil
	})

	// i64.load32_s - load 4 bytes, sign-extend to i64
	I64Load32S = addInstruction(opcodes.I64Load32S, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 4)
		ctx.Stack.Push(int64(int32(binary.LittleEndian.Uint32(bytes))))
		return nil
	})

	// i64.load32_u - load 4 bytes, zero-extend to i64
	I64Load32U = addInstruction(opcodes.I64Load32U, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		base := castInt(ctx.Stack.Pop())
		addr := base + offset
		bytes := ctx.Memories[0].Load(addr, 4)
		ctx.Stack.Push(int64(binary.LittleEndian.Uint32(bytes)))
		return nil
	})

	// i32.store - store i32 as 4 bytes
	I32Store = addInstruction(opcodes.I32Store, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, uint32(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// i64.store - store i64 as 8 bytes
	I64Store = addInstruction(opcodes.I64Store, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt64(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, uint64(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// f32.store - store f32 as 4 bytes
	F32Store = addInstruction(opcodes.F32Store, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := ctx.Stack.Pop().(float32)
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, math.Float32bits(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// f64.store - store f64 as 8 bytes
	F64Store = addInstruction(opcodes.F64Store, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := ctx.Stack.Pop().(float64)
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, math.Float64bits(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// i32.store8 - store low 8 bits of i32
	I32Store8 = addInstruction(opcodes.I32Store8, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		ctx.Memories[0].Store(addr, []byte{byte(value)})

		return nil
	})

	// i32.store16 - store low 16 bits of i32
	I32Store16 = addInstruction(opcodes.I32Store16, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytes, uint16(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// i64.store8 - store low 8 bits of i64
	I64Store8 = addInstruction(opcodes.I64Store8, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt64(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		ctx.Memories[0].Store(addr, []byte{byte(value)})

		return nil
	})

	// i64.store16 - store low 16 bits of i64
	I64Store16 = addInstruction(opcodes.I64Store16, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt64(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(bytes, uint16(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})

	// i64.store32 - store low 32 bits of i64
	I64Store32 = addInstruction(opcodes.I64Store32, func(ctx *execution.Context) error {
		_ = ctx.Body.Varint() // alignment
		offset := ctx.Body.Varint()
		value := castInt64(ctx.Stack.Pop())
		base := castInt(ctx.Stack.Pop())
		addr := base + offset

		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, uint32(value))

		ctx.Memories[0].Store(addr, bytes)

		return nil
	})
)
