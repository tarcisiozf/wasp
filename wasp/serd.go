package wasp

import (
	"encoding/gob"
	"fmt"
	"io"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/memory"
)

type StateStore struct {
	Globals  []GlobalItem
	Memories []MemoryState
	Tables   []TableState
}

type GlobalItem struct {
	Value   any
	Mutable bool
}

type MemoryState struct {
	Data     []byte
	NumPages int
	MaxPages int
}

type TableState = memory.Table

type CallFrame struct {
	FunctionIndex int
	Context       CallContext
}

type BlockFrame struct {
	StartPos int
}

type CallContext struct {
	Stack  []any
	Locals []any

	NumParams  int
	NumResults int
	Params     []any

	BodyPos             int
	BodyCheckpoint      int
	FunctionCallRequest int
	Done                bool
	TailCall            bool

	Condition  bool
	BlockStack []BlockFrame
}

type ExecutionState struct {
	Store     StateStore
	CallStack []CallFrame
}

func SerializeState(dest io.Writer, store *Store, instance *Instance) error {
	stateStore, err := toStateStore(store)
	if err != nil {
		return err
	}
	callStack, err := toCallStack(instance)
	if err != nil {
		return err
	}
	state := ExecutionState{
		Store:     stateStore,
		CallStack: callStack,
	}

	enc := gob.NewEncoder(dest)
	if err := enc.Encode(state); err != nil {
		return fmt.Errorf("failed to encode state: %v", err)
	}

	return nil
}

func DeserializeState(src io.Reader) error {
	var state ExecutionState

	dec := gob.NewDecoder(src)
	if err := dec.Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state: %v", err)
	}

	return nil
}

func toStateStore(store *Store) (StateStore, error) {
	globals, err := toGlobalItems(store.Globals)
	if err != nil {
		return StateStore{}, err
	}
	memories := toMemoryStates(store.Memories)
	tables := toTableStates(store.Tables)
	return StateStore{
		Globals:  globals,
		Memories: memories,
		Tables:   tables,
	}, nil
}

func toGlobalItems(globals *memory.Global) ([]GlobalItem, error) {
	size := globals.Size()
	items := make([]GlobalItem, size)
	for i := 0; i < size; i++ {
		value, mutable, err := globals.Get(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get global at index %d: %v", i, err)
		}
		items[i] = GlobalItem{
			Value:   value,
			Mutable: mutable,
		}
	}
	return items, nil
}

func toMemoryStates(memories []*memory.Memory) []MemoryState {
	items := make([]MemoryState, len(memories))
	for i, mem := range memories {
		items[i] = toMemoryState(mem)
	}
	return items
}

func toMemoryState(mem *memory.Memory) MemoryState {
	return MemoryState{
		Data:     mem.Data(),
		NumPages: mem.NumPages(),
		MaxPages: mem.MaxPages(),
	}
}

func toTableStates(tables []*memory.Table) []TableState {
	items := make([]TableState, len(tables))
	for i, table := range tables {
		items[i] = *table
	}
	return items
}

func toCallStack(instance *Instance) ([]CallFrame, error) {
	size := instance.callStack.Size()
	frames := make([]CallFrame, size)
	for i := 0; i < size; i++ {
		frame := instance.callStack.At(i)
		frames[i] = toCallFrame(frame)
	}
	return frames, nil
}

func toCallFrame(frame *execution.CallFrame) CallFrame {
	stack := make([]any, frame.Context.Stack.Size())
	for i := 0; i < frame.Context.Stack.Size(); i++ {
		stack[i] = frame.Context.Stack.At(i)
	}

	locals := make([]any, frame.Context.Locals.Size())
	for i := 0; i < frame.Context.Locals.Size(); i++ {
		locals[i] = frame.Context.Locals.At(i)
	}

	blockStack := make([]BlockFrame, frame.Context.BlockStack.Size())
	for i := 0; i < frame.Context.BlockStack.Size(); i++ {
		blockFrame := frame.Context.BlockStack.At(i)
		blockStack[i] = BlockFrame{
			StartPos: blockFrame.StartPos,
		}
	}

	return CallFrame{
		FunctionIndex: frame.FunctionIndex,
		Context: CallContext{
			Stack:  stack,
			Locals: locals,

			NumParams:  frame.Context.NumParams,
			NumResults: frame.Context.NumResults,
			Params:     frame.Context.Params,

			BodyPos:             frame.Context.Body.Position(),
			BodyCheckpoint:      frame.Context.Body.Checkpoint(),
			FunctionCallRequest: frame.Context.FunctionCallRequest,
			Done:                frame.Context.Done,
			TailCall:            frame.Context.TailCall,

			Condition:  frame.Context.Condition,
			BlockStack: blockStack,
		},
	}
}
