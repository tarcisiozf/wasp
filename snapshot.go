package wasp

import (
	"fmt"
	"io"

	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/memory"
	"github.com/tarcisiozf/wasp/internal/serialization"
	iface "github.com/tarcisiozf/wasp/memory"
	"github.com/tarcisiozf/wasp/memory/contiguous"
	"github.com/tarcisiozf/wasp/memory/sparse"
)

const (
	tagLinearMemory      byte = 0x01
	tagSparsePagedMemory byte = 0x02
)

func SerializeState(dest serialization.Writer, store *Store, instance *Instance) error {
	encoder := serialization.NewEncoder(dest)

	if err := toStateStore(encoder, store); err != nil {
		return err
	}
	if err := encodeCallStack(encoder, instance); err != nil {
		return err
	}

	return nil
}

func DeserializeState(src io.Reader, module *Module, linker *Linker) (*Store, *Instance, error) {
	decoder := serialization.NewDecoder(src)

	store, err := fromStateStore(decoder)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode store: %w", err)
	}

	instance, err := NewInstance(module, store, WithLinker(linker))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create instance: %w", err)
	}

	if err := decodeCallStack(decoder, module, instance); err != nil {
		return nil, nil, fmt.Errorf("failed to decode call stack: %w", err)
	}

	return store, instance, nil
}

func fromStateStore(decoder *serialization.Decoder) (*Store, error) {
	globals, err := decodeGlobalItems(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode globals: %w", err)
	}
	memories, err := decodeMemories(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode memories: %w", err)
	}
	tables, err := decodeTables(decoder)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tables: %w", err)
	}
	return &Store{
		Globals:  globals,
		Memories: memories,
		Tables:   tables,
	}, nil
}

func decodeGlobalItems(decoder *serialization.Decoder) (*memory.Global, error) {
	size, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	globals := memory.NewGlobal()
	for i := 0; i < size; i++ {
		value, err := decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode global value at index %d: %w", i, err)
		}
		mutable, err := decoder.Bool()
		if err != nil {
			return nil, fmt.Errorf("failed to decode global mutable at index %d: %w", i, err)
		}
		globals.Push(value, mutable)
	}
	return globals, nil
}

func decodeMemories(decoder *serialization.Decoder) ([]iface.Memory, error) {
	count, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	memories := make([]iface.Memory, count)
	for i := 0; i < count; i++ {
		tag, err := decoder.Byte()
		if err != nil {
			return nil, fmt.Errorf("failed to decode memory type tag at index %d: %w", i, err)
		}
		switch tag {
		case tagLinearMemory:
			memories[i], err = decodeLinearMemory(decoder)
			if err != nil {
				return nil, fmt.Errorf("failed to decode linear memory at index %d: %w", i, err)
			}
		case tagSparsePagedMemory:
			memories[i], err = sparse.Decode(decoder)
			if err != nil {
				return nil, fmt.Errorf("failed to decode sparse paged memory at index %d: %w", i, err)
			}
		default:
			return nil, fmt.Errorf("unknown memory type tag 0x%02x at index %d", tag, i)
		}
	}
	return memories, nil
}

func decodeLinearMemory(decoder *serialization.Decoder) (*contiguous.Memory, error) {
	data, err := decoder.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to decode memory data: %w", err)
	}
	numPages, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode memory numPages: %w", err)
	}
	maxPages, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode memory maxPages: %w", err)
	}
	mem := contiguous.NewMemory(numPages, maxPages)
	mem.Store(0, data)
	return mem, nil
}

func decodeTables(decoder *serialization.Decoder) ([]*memory.Table, error) {
	count, err := decoder.Int()
	if err != nil {
		return nil, err
	}
	tables := make([]*memory.Table, count)
	for i := 0; i < count; i++ {
		elementType, err := decoder.Byte()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table elementType at index %d: %w", i, err)
		}
		initialSize, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table initialSize at index %d: %w", i, err)
		}
		maxSize, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table maxSize at index %d: %w", i, err)
		}
		numElements, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode table numElements at index %d: %w", i, err)
		}
		elements := make([]int, numElements)
		for j := 0; j < numElements; j++ {
			elements[j], err = decoder.Int()
			if err != nil {
				return nil, fmt.Errorf("failed to decode table element at index %d/%d: %w", i, j, err)
			}
		}
		tables[i] = &memory.Table{
			ElementType: elementType,
			InitialSize: initialSize,
			MaxSize:     maxSize,
			Elements:    elements,
		}
	}
	return tables, nil
}

func decodeCallStack(decoder *serialization.Decoder, module *Module, instance *Instance) error {
	count, err := decoder.Int()
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		frame, err := decodeCallFrame(decoder, module, instance)
		if err != nil {
			return fmt.Errorf("failed to decode call frame at index %d: %w", i, err)
		}
		instance.callStack.Push(frame)
	}
	return nil
}

func decodeCallFrame(decoder *serialization.Decoder, module *Module, instance *Instance) (*execution.CallFrame, error) {
	// Stack
	stackSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode stack size: %w", err)
	}
	stackItems := make([]any, stackSize)
	for i := 0; i < stackSize; i++ {
		stackItems[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode stack item %d: %w", i, err)
		}
	}

	// Locals
	localsSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode locals size: %w", err)
	}
	localItems := make([]any, localsSize)
	for i := 0; i < localsSize; i++ {
		localItems[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode local item %d: %w", i, err)
		}
	}

	// BlockStack
	blockStackSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode block stack size: %w", err)
	}
	blockStack := memory.NewStackWithCapacity[execution.BlockFrame](blockStackSize)
	for i := 0; i < blockStackSize; i++ {
		startPos, err := decoder.Int()
		if err != nil {
			return nil, fmt.Errorf("failed to decode block frame %d: %w", i, err)
		}
		blockStack.Push(execution.BlockFrame{StartPos: startPos})
	}

	// Body position and checkpoint
	bodyPos, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode body pos: %w", err)
	}
	bodyCheckpoint, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode body checkpoint: %w", err)
	}

	// FunctionIndex
	functionIndex, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode function index: %w", err)
	}

	// NumResults
	numResults, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode num results: %w", err)
	}

	// Params
	paramsSize, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode params size: %w", err)
	}
	params := make([]any, paramsSize)
	for i := 0; i < paramsSize; i++ {
		params[i], err = decoder.Any()
		if err != nil {
			return nil, fmt.Errorf("failed to decode param %d: %w", i, err)
		}
	}

	// FunctionCallRequest, Done, TailCall, Condition
	functionCallRequest, err := decoder.Int()
	if err != nil {
		return nil, fmt.Errorf("failed to decode function call request: %w", err)
	}
	done, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode done: %w", err)
	}
	tailCall, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode tail call: %w", err)
	}
	condition, err := decoder.Bool()
	if err != nil {
		return nil, fmt.Errorf("failed to decode condition: %w", err)
	}

	frame := &execution.CallFrame{
		FunctionIndex: functionIndex,
		Context: execution.Context{
			Memory: instance.memory,
			Locals: memory.NewStack[any](localItems...),

			NumResults: numResults,
			Params:     params,

			FunctionCallRequest: functionCallRequest,
			Done:                done,
			TailCall:            tailCall,

			Condition:  condition,
			BlockStack: blockStack,
		},
	}

	// Resolve the function reference
	if module.IsImport(functionIndex) {
		extFunc, err := instance.getImportedFunc(functionIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to get imported function at index %d: %w", functionIndex, err)
		}
		frame.Function = extFunc
	} else if module.IsFunction(functionIndex) {
		fn := module.FunctionAt(functionIndex)
		frame.Function = fn

		body := binary.NewIterator(fn.Body)
		body.Seek(bodyPos)
		body.SetCheckpointTo(bodyCheckpoint)
		frame.Context.Body = body

		frame.Context.Blocks = fn.Blocks
	}

	return frame, nil
}

func toStateStore(encoder *serialization.Encoder, store *Store) error {
	if err := encodeGlobalItems(encoder, store.Globals); err != nil {
		return err
	}
	encodeMemories(encoder, store.Memories)
	encodeTables(encoder, store.Tables)
	return nil
}

func encodeGlobalItems(encoder *serialization.Encoder, globals *memory.Global) error {
	size := globals.Size()
	encoder.Int(size)
	for i := 0; i < size; i++ {
		value, mutable, err := globals.Get(i)
		if err != nil {
			return fmt.Errorf("failed to get global at index %d: %v", i, err)
		}
		encoder.Any(value)
		encoder.Bool(mutable)
	}
	return nil
}

func encodeMemories(encoder *serialization.Encoder, memories []iface.Memory) {
	encoder.Int(len(memories))
	for _, mem := range memories {
		encodeMemory(encoder, mem)
	}
}

func encodeMemory(encoder *serialization.Encoder, mem any) {
	switch mem.(type) {
	case *contiguous.Memory:
		encoder.Byte(tagLinearMemory)
		encodeLinearMemory(encoder, mem.(*contiguous.Memory))
	case *sparse.Memory:
		encoder.Byte(tagSparsePagedMemory)
		sparse.Encode(encoder, mem.(*sparse.Memory))
	}
}

func encodeLinearMemory(encoder *serialization.Encoder, mem *contiguous.Memory) {
	encoder.Bytes(mem.Data())
	encoder.Int(mem.NumPages())
	encoder.Int(mem.MaxPages())
}

func encodeTables(encoder *serialization.Encoder, tables []*memory.Table) {
	encoder.Int(len(tables))
	for _, table := range tables {
		encoder.Byte(table.ElementType)
		encoder.Int(table.InitialSize)
		encoder.Int(table.MaxSize)
		encoder.Int(len(table.Elements))
		for _, elem := range table.Elements {
			encoder.Int(elem)
		}
	}
}

func encodeCallStack(encoder *serialization.Encoder, instance *Instance) error {
	size := instance.callStack.Size()
	encoder.Int(size)
	for i := 0; i < size; i++ {
		frame := instance.callStack.At(i)
		encodeCallFrame(encoder, frame)
	}
	return nil
}

func encodeCallFrame(encoder *serialization.Encoder, frame *execution.CallFrame) {
	if frame.Context.Stack == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.Stack.Size())
		for i := 0; i < frame.Context.Stack.Size(); i++ {
			encoder.Any(frame.Context.Stack.At(i))
		}
	}

	if frame.Context.Locals == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.Locals.Size())
		for i := 0; i < frame.Context.Locals.Size(); i++ {
			encoder.Any(frame.Context.Locals.At(i))
		}
	}

	if frame.Context.BlockStack == nil {
		encoder.Int(0)
	} else {
		encoder.Int(frame.Context.BlockStack.Size())
		for i := 0; i < frame.Context.BlockStack.Size(); i++ {
			blockFrame := frame.Context.BlockStack.At(i)
			encoder.Int(blockFrame.StartPos)
		}
	}

	var bodyPos, bodyCheckpoint int
	if frame.Context.Body != nil {
		bodyPos = frame.Context.Body.Position()
		bodyCheckpoint = frame.Context.Body.Checkpoint()
	}
	encoder.Int(bodyPos)
	encoder.Int(bodyCheckpoint)

	encoder.Int(frame.FunctionIndex)

	encoder.Int(frame.Context.NumResults)
	encoder.Int(len(frame.Context.Params))
	for _, param := range frame.Context.Params {
		encoder.Any(param)
	}

	encoder.Int(frame.Context.FunctionCallRequest)
	encoder.Bool(frame.Context.Done)
	encoder.Bool(frame.Context.TailCall)
	encoder.Bool(frame.Context.Condition)
}
