package wasp

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/tarcisiozf/wasp/internal/binary"
	execution "github.com/tarcisiozf/wasp/internal/execution"
	memory "github.com/tarcisiozf/wasp/internal/memory"
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

func DeserializeState(src io.Reader, module *Module, linker *Linker) (*Store, *Instance, error) {
	var state ExecutionState

	dec := gob.NewDecoder(src)
	if err := dec.Decode(&state); err != nil {
		return nil, nil, fmt.Errorf("failed to decode state: %v", err)
	}

	store := &Store{
		Globals:  memory.NewGlobal(),
		Memories: make([]*memory.Memory, len(state.Store.Memories)),
		Tables:   make([]*memory.Table, len(state.Store.Tables)),
	}
	for _, item := range state.Store.Globals {
		store.Globals.Push(item.Value, item.Mutable)
	}
	for i, memState := range state.Store.Memories {
		mem := memory.NewMemory(memState.NumPages, memState.MaxPages)
		mem.Store(0, memState.Data)
		store.Memories[i] = mem
	}
	for i, table := range state.Store.Tables {
		store.Tables[i] = &memory.Table{
			ElementType: table.ElementType,
			InitialSize: table.InitialSize,
			MaxSize:     table.MaxSize,
			Elements:    table.Elements,
		}
	}

	instance, err := NewInstance(module, store, WithLinker(linker))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create instance: %v", err)
	}

	for _, frameState := range state.CallStack {
		frame := &execution.CallFrame{
			FunctionIndex: frameState.FunctionIndex,
			Context: execution.Context{
				Stack:  memory.NewStack[any](frameState.Context.Stack...),
				Locals: memory.NewStack[any](frameState.Context.Locals...),

				NumParams:  frameState.Context.NumParams,
				NumResults: frameState.Context.NumResults,
				Params:     frameState.Context.Params,

				FunctionCallRequest: frameState.Context.FunctionCallRequest,
				Done:                frameState.Context.Done,
				TailCall:            frameState.Context.TailCall,

				Condition: frameState.Context.Condition,
			},
		}

		frame.Context.Globals = store.Globals
		frame.Context.Memories = store.Memories
		frame.Context.Tables = store.Tables

		frame.Context.FuncSignatures = instance.funcSignatures
		frame.Context.TypeSignatures = instance.typeSignatures

		if module.IsImport(frameState.FunctionIndex) {
			extFunc, err := instance.getImportedFunc(frameState.FunctionIndex)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get imported function at index %d: %v", frameState.FunctionIndex, err)
			}
			frame.Function = extFunc
		}

		if module.IsFunction(frameState.FunctionIndex) {
			fn := module.FunctionAt(frameState.FunctionIndex)
			frame.Function = fn

			body := binary.NewIterator(fn.Body)
			body.Seek(frameState.Context.BodyPos)
			body.SetCheckpointTo(frameState.Context.BodyCheckpoint)
			frame.Context.Body = body

			frame.Context.Blocks = fn.Blocks
		}

		frame.Context.BlockStack = memory.NewStack[execution.BlockFrame]()
		for _, blockFrame := range frameState.Context.BlockStack {
			frame.Context.BlockStack.Push(execution.BlockFrame{
				StartPos: blockFrame.StartPos,
			})
		}

		instance.callStack.Push(frame)
	}

	return store, instance, nil
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
	var stack []any
	if frame.Context.Stack != nil {
		stack = make([]any, frame.Context.Stack.Size())
		for i := 0; i < frame.Context.Stack.Size(); i++ {
			stack[i] = frame.Context.Stack.At(i)
		}
	}

	var locals []any
	if frame.Context.Locals != nil {
		locals = make([]any, frame.Context.Locals.Size())
		for i := 0; i < frame.Context.Locals.Size(); i++ {
			locals[i] = frame.Context.Locals.At(i)
		}
	}

	var blockStack []BlockFrame
	if frame.Context.BlockStack != nil {
		blockStack = make([]BlockFrame, frame.Context.BlockStack.Size())
		for i := 0; i < frame.Context.BlockStack.Size(); i++ {
			blockFrame := frame.Context.BlockStack.At(i)
			blockStack[i] = BlockFrame{
				StartPos: blockFrame.StartPos,
			}
		}
	}

	var bodyPos, bodyCheckpoint int
	if frame.Context.Body != nil {
		bodyPos = frame.Context.Body.Position()
		bodyCheckpoint = frame.Context.Body.Checkpoint()
	}

	return CallFrame{
		FunctionIndex: frame.FunctionIndex,
		Context: CallContext{
			Stack:  stack,
			Locals: locals,

			NumParams:  frame.Context.NumParams,
			NumResults: frame.Context.NumResults,
			Params:     frame.Context.Params,

			BodyPos:             bodyPos,
			BodyCheckpoint:      bodyCheckpoint,
			FunctionCallRequest: frame.Context.FunctionCallRequest,
			Done:                frame.Context.Done,
			TailCall:            frame.Context.TailCall,

			Condition:  frame.Context.Condition,
			BlockStack: blockStack,
		},
	}
}
