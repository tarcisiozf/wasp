package wasp

import (
	"encoding/binary"
	"fmt"
	"io"
	biniter "wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/memory"
)

type itemType byte

const (
	typeInt32 itemType = iota + 1
	typeInt64
	typeFloat32
	typeFloat64
)

func Foo(dest io.Writer, store *Store, instance *Instance) error {
	marshalStoreGlobals(dest, store.Globals)
	marshalStoreMemories(dest, store.Memories)
	marshalStoreTables(dest, store.Tables)
	marshalStack(dest, instance.callStack)
	return nil
}

func marshalStoreGlobals(dest io.Writer, globals *memory.Global) {
	size := globals.Size()
	marshalInt(dest, size)
	for i := 0; i < size; i++ {
		value, mutable, err := globals.Get(i)
		if err != nil {
			panic(fmt.Sprintf("failed to get global at index %d: %v", i, err))
		}
		marshal(dest, mutable)
		marshal(dest, value)
	}
}

func marshalStoreMemories(dest io.Writer, memories []*memory.Memory) {
}

func marshalStoreTables(dest io.Writer, tables []*memory.Table) {

}

func marshalStack[T any](dest io.Writer, stack *memory.Stack[T]) {
	size := stack.Size()
	marshalInt(dest, size)
	for i := 0; i < size; i++ {
		marshal(dest, stack.At(i))
	}
}

func marshalSlice[T any](dest io.Writer, slice []T) {
	size := len(slice)
	marshalInt(dest, size)
	for i := 0; i < size; i++ {
		marshal(dest, slice[i])
	}
}

func marshalCallFrame(dest io.Writer, frame *execution.CallFrame) {
	marshalInt(dest, frame.FunctionIndex)

	marshalStack(dest, frame.Context.Locals)
	marshalStack(dest, frame.Context.Stack)

	marshal(dest, frame.Context.NumParams)
	marshal(dest, frame.Context.NumResults)
	marshalSlice(dest, frame.Context.Params)

	marshalIterator(dest, frame.Context.Body)
	marshal(dest, frame.Context.FunctionCallRequest)
	marshal(dest, frame.Context.Done)
	marshal(dest, frame.Context.TailCall)

	marshal(dest, frame.Context.Condition)
	marshalStack(dest, frame.Context.BlockStack)
}

func marshalIterator(dest io.Writer, iter *biniter.Iterator) {
	marshalInt(dest, iter.Position())
	marshalInt(dest, iter.Checkpoint())
}

func marshalInt(dest io.Writer, count int) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(count))
	write(dest, b...)
}

func marshalBool(dest io.Writer, cond bool) {
	b := byte(0)
	if cond {
		b = 1
	}
	write(dest, b)
}

func marshal(dest io.Writer, item any) {
	switch item.(type) {
	case *execution.CallFrame:
		marshalCallFrame(dest, item.(*execution.CallFrame))
	case int:
		marshalInt(dest, item.(int))
	case bool:
		marshalBool(dest, item.(bool))
	case int32:
		marshalInt32(dest, item.(int32))
	case execution.BlockFrame:
		marshalBlockFrame(dest, item.(execution.BlockFrame))
	default:
		panic(fmt.Sprintf("unsupported type: %T", item))
	}
}

func marshalBlockFrame(dest io.Writer, blockFrame execution.BlockFrame) {
	marshalInt(dest, blockFrame.StartPos)
}

func marshalInt32(dest io.Writer, value int32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(value))
	write(dest, b...)
}

func write(dest io.Writer, b ...byte) {
	n, err := dest.Write(b)
	if err != nil {
		panic(fmt.Sprintf("failed to write: %v", err))
	}
	if n != len(b) {
		panic(fmt.Sprintf("failed to write all bytes: %d of %d", n, len(b)))
	}
}
