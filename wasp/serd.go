package wasp

import (
	"encoding/gob"
	"fmt"
	"io"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/memory"
)

type fooStore struct {
	Globals  []fooGlobalItem
	Memories []fooMemory
	Tables   []fooTables
}

type fooGlobalItem struct {
	Value   any
	Mutable bool
}

type fooMemory struct {
	Data     []byte
	NumPages int
	MaxPages int
}

type fooTables = memory.Table

type fooCallFrame struct {
	FunctionIndex int
	Context       fooCallContext
}

type fooBlockFrame struct {
	StartPos int
}

type fooCallContext struct {
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
	BlockStack []fooBlockFrame
}

type fooState struct {
	Store     fooStore
	CallStack []fooCallFrame
}

func Foo(dest io.Writer, store *Store, instance *Instance) error {
	fooStore, err := toFooStore(store)
	if err != nil {
		return err
	}
	callStack, err := toFooCallStack(instance)
	if err != nil {
		return err
	}
	state := fooState{
		Store:     fooStore,
		CallStack: callStack,
	}

	enc := gob.NewEncoder(dest)
	if err := enc.Encode(state); err != nil {
		return fmt.Errorf("failed to encode state: %v", err)
	}

	return nil
}

func Bar(src io.Reader) error {
	var state fooState

	dec := gob.NewDecoder(src)
	if err := dec.Decode(&state); err != nil {
		return fmt.Errorf("failed to decode state: %v", err)
	}

	return nil
}

func toFooStore(store *Store) (fooStore, error) {
	globals, err := toFooGlobals(store.Globals)
	if err != nil {
		return fooStore{}, err
	}
	memories := toFooMemories(store.Memories)
	tables := toFooTables(store.Tables)
	return fooStore{
		Globals:  globals,
		Memories: memories,
		Tables:   tables,
	}, nil
}

func toFooGlobals(globals *memory.Global) ([]fooGlobalItem, error) {
	size := globals.Size()
	items := make([]fooGlobalItem, size)
	for i := 0; i < size; i++ {
		value, mutable, err := globals.Get(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get global at index %d: %v", i, err)
		}
		items[i] = fooGlobalItem{
			Value:   value,
			Mutable: mutable,
		}
	}
	return items, nil
}

func toFooMemories(memories []*memory.Memory) []fooMemory {
	items := make([]fooMemory, len(memories))
	for i, mem := range memories {
		items[i] = toFooMemory(mem)
	}
	return items
}

func toFooMemory(mem *memory.Memory) fooMemory {
	return fooMemory{
		Data:     mem.Data(),
		NumPages: mem.NumPages(),
		MaxPages: mem.MaxPages(),
	}
}

func toFooTables(tables []*memory.Table) []fooTables {
	items := make([]fooTables, len(tables))
	for i, table := range tables {
		items[i] = *table
	}
	return items
}

func toFooCallStack(instance *Instance) ([]fooCallFrame, error) {
	size := instance.callStack.Size()
	frames := make([]fooCallFrame, size)
	for i := 0; i < size; i++ {
		frame := instance.callStack.At(i)
		frames[i] = toFooCallFrame(frame)
	}
	return frames, nil
}

func toFooCallFrame(frame *execution.CallFrame) fooCallFrame {
	stack := make([]any, frame.Context.Stack.Size())
	for i := 0; i < frame.Context.Stack.Size(); i++ {
		stack[i] = frame.Context.Stack.At(i)
	}

	locals := make([]any, frame.Context.Locals.Size())
	for i := 0; i < frame.Context.Locals.Size(); i++ {
		locals[i] = frame.Context.Locals.At(i)
	}

	blockStack := make([]fooBlockFrame, frame.Context.BlockStack.Size())
	for i := 0; i < frame.Context.BlockStack.Size(); i++ {
		blockFrame := frame.Context.BlockStack.At(i)
		blockStack[i] = fooBlockFrame{
			StartPos: blockFrame.StartPos,
		}
	}

	return fooCallFrame{
		FunctionIndex: frame.FunctionIndex,
		Context: fooCallContext{
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
