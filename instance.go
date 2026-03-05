package wasp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/external"
	"github.com/tarcisiozf/wasp/internal/memory"
	"github.com/tarcisiozf/wasp/internal/module"
)

const (
	maxCallStackDepth = 1024
)

type InstanceOption func(*Instance) error

func WithLinker(linker *Linker) InstanceOption {
	return func(e *Instance) error {
		e.linker = linker
		return nil
	}
}

func VerboseShowFunctionCalls() InstanceOption {
	return func(e *Instance) error {
		e.debug.showFunctionCallsFlag = true
		return nil
	}
}

func VerboseShowImports() InstanceOption {
	return func(e *Instance) error {
		e.debug.showImportsFlag = true
		return nil
	}
}

func VerboseShowExports() InstanceOption {
	return func(e *Instance) error {
		e.debug.showExportsFlag = true
		return nil
	}
}

func VerboseShowAssembly() InstanceOption {
	return func(e *Instance) error {
		e.debug.showAssemblyFlag = true
		return nil
	}
}

func Verbose() InstanceOption {
	return func(e *Instance) error {
		e.debug.showFunctionCallsFlag = true
		e.debug.showImportsFlag = true
		e.debug.showExportsFlag = true
		e.debug.showAssemblyFlag = true
		e.debug.showInstructions = true
		e.debug.showCustomSection = true
		return nil
	}
}

// IgnoreUnreachable makes the interpreter continue past unreachable instructions
// instead of returning an error. Useful when running WASM with UBSan that panics
// on undefined behavior but you want to continue anyway.
func IgnoreUnreachable() InstanceOption {
	return func(e *Instance) error {
		e.ignoreUnreachable = true
		return nil
	}
}

type DebugData struct {
	showFunctionCallsFlag bool
	showImportsFlag       bool
	showExportsFlag       bool
	showAssemblyFlag      bool
	showInstructions      bool
	showCustomSection     bool

	modules   []string
	functions []string
}

type Instance struct {
	module *module.Module
	linker *Linker
	store  *Store

	indexedImportedFunctions []*external.Function
	debug                    DebugData
	ignoreUnreachable        bool
	paused                   bool

	memory *execution.Memory
	// Keep track of the original/root frame for tail calls
	rootFrame *execution.CallFrame
	callStack *memory.Stack[*execution.CallFrame]
}

func NewInstance(module *module.Module, store *Store, options ...InstanceOption) (*Instance, error) {
	funcSignatures := module.FunctionSignatures()
	typeSignatures := module.TypeSignatures()

	callStack := memory.NewStackWithCapacity[*execution.CallFrame](64)

	stack := memory.NewStackWithCapacity[any](32)

	memory := &execution.Memory{
		Stack:          stack,
		Globals:        store.Globals,
		Memories:       store.Memories,
		Tables:         store.Tables,
		FuncSignatures: funcSignatures,
		TypeSignatures: typeSignatures,
	}

	instance := &Instance{
		module: module,
		store:  store,

		memory:    memory,
		callStack: callStack,
	}
	for _, option := range options {
		if err := option(instance); err != nil {
			return nil, fmt.Errorf("failed to apply instance option: %w", err)
		}
	}

	if instance.linker == nil {
		instance.linker = NewLinker()
	}

	//if instance.debug.showAssemblyFlag {
	//	if err := debug.WasmToString(os.Stdout, module.Wasm()); err != nil {
	//		return nil, fmt.Errorf("failed to disassemble module: %w", err)
	//	}
	//}

	if instance.debug.showImportsFlag {
		imports := module.Imports()
		fmt.Printf("Imports (count %d):\n", len(imports))
		for _, imp := range imports {
			fmt.Printf("\t%s\n", imp.String())
		}
	}

	if instance.debug.showExportsFlag {
		exports := module.Exports()
		fmt.Printf("Exports (count %d):\n", len(exports))
		for _, exp := range exports {
			fmt.Printf("\t%s\n", exp.String())
		}
	}

	if instance.debug.showCustomSection {
		sections := module.CustomSections()
		fmt.Printf("Custom Sections (count %d):\n", len(sections))
		for name, data := range sections {
			fmt.Printf("\t%s (size %d bytes)\n", name, len(data))
		}
	}

	if err := instance.mapImportsToExternalFunctions(); err != nil {
		return nil, fmt.Errorf("invalid imports: %w", err)
	}

	return instance, nil
}

func (instance *Instance) Call(fnIndex int, params ...any) (*execution.CallFrame, error) {
	if !instance.module.IsFunction(fnIndex) {
		return nil, fmt.Errorf("invalid function index: %d", fnIndex)
	}
	fn := instance.module.FunctionAt(fnIndex)
	if len(params) != len(fn.Signature.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.Signature.Params), len(params))
	}
	instance.memory.Stack.Push(params...)
	return instance.enqueueCall(fnIndex)
}

func (instance *Instance) enqueueCall(fnIndex int) (*execution.CallFrame, error) {
	if instance.callStack.Size()+1 > maxCallStackDepth {
		return nil, fmt.Errorf("call stack stack overflow")
	}

	callFrame, err := instance.createCallFrame(fnIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to create enqueue call frame: %w", err)
	}

	instance.callStack.Push(callFrame)

	return callFrame, nil
}

func (instance *Instance) Run() error {
	if instance.paused {
		instance.setPauseState(false)
	}

	for !instance.Done() && !instance.paused {
		if err := instance.Tick(); err != nil {
			return fmt.Errorf("error during execution: %w", err)
		}
	}

	return nil
}

func (instance *Instance) Tick() error {
	callFrame := instance.callStack.Top()

	if callFrame.Done() {
		instance.callStack.Pop()

		// forward call results to previous frame if it exists
		prev := instance.callStack.Top()
		if prev == nil && instance.rootFrame != nil {
			// No previous frame but we have a root frame from a tail call chain
			// Copy results to the root frame
			instance.rootFrame.Context.Done = true
			prev = instance.rootFrame
		}

		//if prev != nil {
		//	results := callFrame.Context.Results()

		//if instance.debug.showFunctionCallsFlag {
		//	fmt.Printf("\tresults: [%s]\n\n", formatArgs(results))
		//}

		//prev.Context.Stack.Push(results...)
		//}

		return nil
	}

	if instance.debug.showFunctionCallsFlag {
		fnindex := callFrame.FunctionIndex
		fntype := '@'
		if fnindex >= len(instance.indexedImportedFunctions) {
			fntype = '$'
			fnindex = fnindex - len(instance.indexedImportedFunctions)
		}
		fnname := callFrame.Function.String()
		fmt.Printf("%c%s ( %s ) - %d (0x%x)\n", fntype, fnname, formatArgs(callFrame.Context.Params), fnindex, fnindex)
	}

	if err := callFrame.Call(); err != nil {
		// Check if this is an unreachable error and we're configured to ignore it
		if instance.ignoreUnreachable && errors.Is(err, execution.ErrUnreachable) {
			// Mark frame as done and continue - this will pop it and return to caller
			callFrame.Context.Done = true
			return nil
		}
		return fmt.Errorf("error enqueueing call frame: %w", err)
	}

	if callFrame.Context.FunctionCallRequest >= 0 {
		fnIndex := callFrame.Context.FunctionCallRequest
		callFrame.Context.FunctionCallRequest = -1

		if callFrame.Context.TailCall {
			// For tail calls, remember the root frame if this is the first in the chain
			if instance.rootFrame == nil {
				instance.rootFrame = callFrame
			}
			instance.callStack.Pop()
			callFrame.Context.TailCall = false
		}

		if _, err := instance.enqueueCall(fnIndex); err != nil {
			return fmt.Errorf("failed to enqueue call function at index %d: %w", fnIndex, err)
		}
	}

	return nil
}

func (instance *Instance) mapImportsToExternalFunctions() error {
	imports := instance.module.Imports()
	instance.indexedImportedFunctions = make([]*external.Function, len(imports))

	for i, imp := range imports {
		extFunc, err := instance.linker.Get(imp.ModuleName, imp.FieldName)
		if err != nil {
			return fmt.Errorf("import %s.%s not found: %w", imp.ModuleName, imp.FieldName, err)
		}
		if err := extFunc.CheckSignatureCompatibility(imp.Signature); err != nil {
			return fmt.Errorf("import %s.%s has incompatible signature: %w", imp.ModuleName, imp.FieldName, err)
		}
		instance.indexedImportedFunctions[i] = extFunc
	}

	return nil
}

func (instance *Instance) getImportedFunc(index int) (*external.Function, error) {
	if index < 0 || index >= len(instance.indexedImportedFunctions) {
		return nil, fmt.Errorf("import index %d out of bounds", index)
	}
	return instance.indexedImportedFunctions[index], nil
}

func (instance *Instance) numParamsForFunc(index int) (int, error) {
	if instance.module.IsImport(index) {
		extFunc, err := instance.getImportedFunc(index)
		if err != nil {
			return -1, fmt.Errorf("invalid function call request: %w", err)
		}
		return extFunc.NumInputs(), nil
	}
	if instance.module.IsFunction(index) {
		fn := instance.module.FunctionAt(index)
		return len(fn.Signature.Params), nil
	}
	return -1, fmt.Errorf("invalid function index: %d", index)
}

func (instance *Instance) createCallFrame(index int) (*execution.CallFrame, error) {
	if instance.module.IsFunction(index) {
		return instance.createLocalCallFrame(index)
	}
	if instance.module.IsImport(index) {
		return instance.createImportCallFrame(index)
	}
	return nil, fmt.Errorf("invalid function index: %d", index)
}

func (instance *Instance) createImportCallFrame(index int) (*execution.CallFrame, error) {
	extFunc, err := instance.getImportedFunc(index)
	if err != nil {
		return nil, fmt.Errorf("invalid function enqueueCall request: %w", err)
	}

	numParams := extFunc.NumInputs()
	numResults := extFunc.NumOutputs()

	if instance.memory.Stack.Size() < numParams {
		return nil, fmt.Errorf("not enough parameters on stack for function at index %d: expected %d, got %d", index, numParams, instance.memory.Stack.Size())
	}
	params := instance.memory.Stack.Last(numParams)

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      extFunc,

		Context: execution.Context{
			Memory: instance.memory,

			NumResults: numResults,
			Params:     params,

			FunctionCallRequest: -1,
		},
	}, nil
}

func (instance *Instance) createLocalCallFrame(index int) (*execution.CallFrame, error) {
	fn := instance.module.FunctionAt(index)

	numParams := len(fn.Signature.Params)
	numResults := len(fn.Signature.Results)

	if instance.memory.Stack.Size() < numParams {
		return nil, fmt.Errorf("not enough parameters on stack for function at index %d: expected %d, got %d", index, numParams, instance.memory.Stack.Size())
	}
	params := instance.memory.Stack.Last(numParams)

	debugEnabled := instance.debug.showInstructions

	locals := memory.NewStackWithCapacity[any](numParams + len(fn.Locals))
	locals.Push(params...)
	locals.Push(fn.Locals...)

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      fn,

		Context: execution.Context{
			Memory: instance.memory,

			NumResults: numResults,
			Params:     params,

			Locals: locals,

			Body:                binary.NewIterator(fn.Body),
			FunctionCallRequest: -1,
			BlockStack:          memory.NewStack[execution.BlockFrame](),
			Blocks:              fn.Blocks,

			Debug: debugEnabled,
		},
	}, nil
}

func (instance *Instance) Pause() {
	instance.setPauseState(true)
}

func (instance *Instance) setPauseState(paused bool) {
	instance.paused = paused
	for callFrame := range instance.callStack.Iter() {
		callFrame.Context.Paused = paused
	}
}

func (instance *Instance) Done() bool {
	return instance.callStack.IsEmpty()
}

func formatArgs(list []any) string {
	out := make([]string, len(list))
	for i, param := range list {
		out[i] = fmt.Sprintf("%T(%v)", param, param)
	}
	return strings.Join(out, ", ")
}
