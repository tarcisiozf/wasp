package instructions_test

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/tarcisiozf/wasp/internal/binary/leb"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/instructions"
	iface "github.com/tarcisiozf/wasp/memory"
	"github.com/tarcisiozf/wasp/memory/contiguous"

	"github.com/stretchr/testify/assert"
)

// createTestContextWithMemory creates a test context with a memory instance
func createTestContextWithMemory(numPages int, body ...[]byte) *execution.Context {
	ctx := createTestContext(body...)
	mem := contiguous.NewMemory(numPages, 10) // max 10 pages
	ctx.Memories = []iface.Memory{mem}
	return ctx
}

func TestI32Load(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store a value in memory first
	value := uint32(0x12345678)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	ctx.Memories[0].Store(8, bytes)

	// Push base address
	ctx.Stack.Push(int32(8))

	err := instructions.I32Load.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(value), ctx.Stack.Top())
}

func TestI32LoadWithOffset(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0),  // alignment
		leb.Encode(10), // offset
	)

	// Store a value in memory at position 18 (base 8 + offset 10)
	value := uint32(0xDEADBEEF)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	ctx.Memories[0].Store(18, bytes)

	// Push base address
	ctx.Stack.Push(int32(8))

	err := instructions.I32Load.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(value), ctx.Stack.Top())
}

func TestI32Store(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int32(0x12345678)
	// Push base address first, then value (stack order)
	ctx.Stack.Push(int32(16)) // base address
	ctx.Stack.Push(value)     // value to store

	err := instructions.I32Store.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents
	bytes := ctx.Memories[0].Load(16, 4)
	result := binary.LittleEndian.Uint32(bytes)
	assert.Equal(t, uint32(value), result)
}

func TestI64LoadUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store a value in memory first
	value := uint64(0x123456789ABCDEF0)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)
	ctx.Memories[0].Store(8, bytes)

	// Push base address
	ctx.Stack.Push(int32(8))

	err := instructions.I64Load.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int64(value), ctx.Stack.Top())
}

func TestI64StoreUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int64(0x123456789ABCDEF0)
	ctx.Stack.Push(int32(16)) // base address
	ctx.Stack.Push(value)     // value to store

	err := instructions.I64Store.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents
	bytes := ctx.Memories[0].Load(16, 8)
	result := binary.LittleEndian.Uint64(bytes)
	assert.Equal(t, uint64(value), result)
}

func TestF32StoreUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := float32(3.14159)
	ctx.Stack.Push(int32(16)) // base address
	ctx.Stack.Push(value)     // value to store

	err := instructions.F32Store.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents
	bytes := ctx.Memories[0].Load(16, 4)
	result := math.Float32frombits(binary.LittleEndian.Uint32(bytes))
	assert.Equal(t, value, result)
}

func TestF64StoreUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := float64(3.141592653589793)
	ctx.Stack.Push(int32(16)) // base address
	ctx.Stack.Push(value)     // value to store

	err := instructions.F64Store.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents
	bytes := ctx.Memories[0].Load(16, 8)
	result := math.Float64frombits(binary.LittleEndian.Uint64(bytes))
	assert.Equal(t, value, result)
}

func TestMemoryGrowUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // memory index
	)

	initialPages := ctx.Memories[0].NumPages()
	ctx.Stack.Push(int32(2)) // grow by 2 pages

	err := instructions.MemoryGrow.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, initialPages, ctx.Stack.Top()) // returns previous size
	assert.Equal(t, initialPages+2, ctx.Memories[0].NumPages())
}

func TestMemorySizeUnit(t *testing.T) {
	ctx := createTestContextWithMemory(3,
		leb.Encode(0), // memory index
	)

	err := instructions.MemorySize.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, 3, ctx.Stack.Top())
}

func TestI32Store8Unit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int32(0x12345678) // only low byte (0x78) should be stored
	ctx.Stack.Push(int32(16))  // base address
	ctx.Stack.Push(value)      // value to store

	err := instructions.I32Store8.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents - only 1 byte stored
	bytes := ctx.Memories[0].Load(16, 1)
	assert.Equal(t, byte(0x78), bytes[0])
}

func TestI32Store16Unit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int32(0x12345678) // only low 2 bytes (0x5678) should be stored
	ctx.Stack.Push(int32(16))  // base address
	ctx.Stack.Push(value)      // value to store

	err := instructions.I32Store16.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents - only 2 bytes stored
	bytes := ctx.Memories[0].Load(16, 2)
	result := binary.LittleEndian.Uint16(bytes)
	assert.Equal(t, uint16(0x5678), result)
}

func TestI64Store8Unit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int64(0x123456789ABCDEF0) // only low byte (0xF0) should be stored
	ctx.Stack.Push(int32(16))          // base address
	ctx.Stack.Push(value)              // value to store

	err := instructions.I64Store8.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents - only 1 byte stored
	bytes := ctx.Memories[0].Load(16, 1)
	assert.Equal(t, byte(0xF0), bytes[0])
}

func TestI64Store16Unit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int64(0x123456789ABCDEF0) // only low 2 bytes (0xDEF0) should be stored
	ctx.Stack.Push(int32(16))          // base address
	ctx.Stack.Push(value)              // value to store

	err := instructions.I64Store16.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents - only 2 bytes stored
	bytes := ctx.Memories[0].Load(16, 2)
	result := binary.LittleEndian.Uint16(bytes)
	assert.Equal(t, uint16(0xDEF0), result)
}

func TestI64Store32Unit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := int64(0x123456789ABCDEF0) // only low 4 bytes (0x9ABCDEF0) should be stored
	ctx.Stack.Push(int32(16))          // base address
	ctx.Stack.Push(value)              // value to store

	err := instructions.I64Store32.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents - only 4 bytes stored
	bytes := ctx.Memories[0].Load(16, 4)
	result := binary.LittleEndian.Uint32(bytes)
	assert.Equal(t, uint32(0x9ABCDEF0), result)
}

func TestI32Load8SUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store a negative value as signed byte (-50 = 0xCE)
	ctx.Memories[0].Store(8, []byte{0xCE})

	ctx.Stack.Push(int32(8))

	err := instructions.I32Load8S.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(-50), ctx.Stack.Top()) // sign-extended
}

func TestI32Load8UUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store 0xCE (206 unsigned, -50 signed)
	ctx.Memories[0].Store(8, []byte{0xCE})

	ctx.Stack.Push(int32(8))

	err := instructions.I32Load8U.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(206), ctx.Stack.Top()) // zero-extended
}

func TestI32Load16SUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store a negative value as signed 16-bit (-1000 = 0xFC18)
	bytes := make([]byte, 2)
	signedVal := int16(-1000)
	binary.LittleEndian.PutUint16(bytes, uint16(signedVal))
	ctx.Memories[0].Store(8, bytes)

	ctx.Stack.Push(int32(8))

	err := instructions.I32Load16S.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(-1000), ctx.Stack.Top()) // sign-extended
}

func TestI32Load16UUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	// Store 0xFC18 (64536 unsigned, -1000 signed)
	bytes := make([]byte, 2)
	signedVal := int16(-1000)
	binary.LittleEndian.PutUint16(bytes, uint16(signedVal))
	ctx.Memories[0].Store(8, bytes)

	ctx.Stack.Push(int32(8))

	err := instructions.I32Load16U.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, int32(64536), ctx.Stack.Top()) // zero-extended
}

func TestF32LoadUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := float32(3.14159)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, math.Float32bits(value))
	ctx.Memories[0].Store(8, bytes)

	ctx.Stack.Push(int32(8))

	err := instructions.F32Load.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, value, ctx.Stack.Top())
}

func TestF64LoadUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // alignment
		leb.Encode(0), // offset
	)

	value := float64(3.141592653589793)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(value))
	ctx.Memories[0].Store(8, bytes)

	ctx.Stack.Push(int32(8))

	err := instructions.F64Load.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, ctx.Stack.Size())
	assert.Equal(t, value, ctx.Stack.Top())
}

func TestMemoryFillUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // memory index
	)

	// Push offset, value, size (in stack order: size, value, offset)
	ctx.Stack.Push(int32(10))   // offset (destination)
	ctx.Stack.Push(int32(0xAB)) // value to fill
	ctx.Stack.Push(int32(5))    // size

	err := instructions.MemoryFill.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify memory contents
	bytes := ctx.Memories[0].Load(10, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, byte(0xAB), bytes[i])
	}
}

func TestMemoryCopyUnit(t *testing.T) {
	ctx := createTestContextWithMemory(1,
		leb.Encode(0), // dst memory index
		leb.Encode(0), // src memory index
	)

	// Store some data at source location
	srcData := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	ctx.Memories[0].Store(100, srcData)

	// Push dst offset, src offset, size (in stack order: size, src, dst)
	ctx.Stack.Push(int32(200)) // dst offset
	ctx.Stack.Push(int32(100)) // src offset
	ctx.Stack.Push(int32(5))   // size

	err := instructions.MemoryCopy.Handler(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, ctx.Stack.Size())

	// Verify copied data
	bytes := ctx.Memories[0].Load(200, 5)
	assert.Equal(t, srcData, bytes)
}
