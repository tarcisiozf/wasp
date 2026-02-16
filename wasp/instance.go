package wasp

import (
	"fmt"
	"math"
	"os"
	"wasp/wasp/debug"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/external"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/module"
)

const (
	maxCallStackDepth = 1024

	verboseShowFunctionCallsFlag = 1 << 0
	verboseShowImportsFlag       = 1 << 1
	verboseShowExportsFlag       = 1 << 2
	verboseShowAssemblyFlag      = 1 << 3
	verboseShowInstructions      = 1 << 4
	verboseShowCustomSection     = 1 << 5
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
		e.debug.flags |= verboseShowFunctionCallsFlag
		return nil
	}
}

func VerboseShowImports() InstanceOption {
	return func(e *Instance) error {
		e.debug.flags |= verboseShowImportsFlag
		return nil
	}
}

func VerboseShowExports() InstanceOption {
	return func(e *Instance) error {
		e.debug.flags |= verboseShowExportsFlag
		return nil
	}
}

func VerboseShowAssembly() InstanceOption {
	return func(e *Instance) error {
		e.debug.flags |= verboseShowAssemblyFlag
		return nil
	}
}

func Verbose() InstanceOption {
	return func(e *Instance) error {
		e.debug.flags = math.MaxUint64
		return nil
	}
}

type DebugData struct {
	flags     uint64
	modules   []string
	functions []string
}

type Instance struct {
	module         *module.Module
	funcSignatures []fnsig.Signature

	linker *Linker
	store  *Store

	indexedImportedFunctions []*external.Function
	debug                    DebugData

	callStack *memory.Stack[*execution.CallFrame]
}

func NewInstance(module *module.Module, store *Store, options ...InstanceOption) (*Instance, error) {
	funcSignatures := module.FunctionSignatures()

	callStack := memory.NewStack[*execution.CallFrame]()

	instance := &Instance{
		module:         module,
		funcSignatures: funcSignatures,

		store: store,

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

	if instance.debug.flags&verboseShowAssemblyFlag != 0 {
		if err := debug.WasmToString(os.Stdout, module.Wasm()); err != nil {
			return nil, fmt.Errorf("failed to disassemble module: %w", err)
		}
	}

	if instance.debug.flags&verboseShowImportsFlag != 0 {
		imports := module.Imports()
		fmt.Printf("Imports (count %d):\n", len(imports))
		for _, imp := range imports {
			fmt.Printf("\t%s\n", imp.String())
		}
	}

	if instance.debug.flags&verboseShowExportsFlag != 0 {
		exports := module.Exports()
		fmt.Printf("Exports (count %d):\n", len(exports))
		for _, exp := range exports {
			fmt.Printf("\t%s\n", exp.String())
		}
	}

	if instance.debug.flags&verboseShowCustomSection != 0 {
		sections := module.CustomSections()
		fmt.Printf("Custom Sections (count %d):\n", len(sections))
		for name, data := range sections {
			fmt.Printf("\t%s (size %d bytes)\n", name, len(data))
		}
	}

	if instance.debug.flags&verboseShowFunctionCallsFlag != 0 {
		sections := module.CustomSections()
		section, ok := sections["name"]
		if ok {
			iter := binary.NewIterator(section)
			for iter.HasNext() {
				subID := iter.Byte()
				subSize := iter.Varint()

				switch subID {
				case 0x00: // module name
					name := iter.String(iter.Varint())
					instance.debug.modules = append(instance.debug.modules, name)
				case 0x01: // function names
					count := iter.Varint()
					instance.debug.functions = make([]string, count)
					for i := 0; i < count; i++ {
						index := iter.Varint()
						name := iter.String(iter.Varint())
						instance.debug.functions[index] = name
					}
				//case 0x02: // local names
				//	funcIndex := iter.Varint()
				//	count := iter.Varint()
				//	bar.locals = make([]string, count)
				//	for i := 0; i < count; i++ {
				//		index := iter.Varint()
				//		name := iter.String(iter.Varint())
				//		bar.locals[index] = name
				//	}
				default:
					// Skip unknown subsection
					fmt.Printf("\tUnknown Subsection ID: 0x%x (skipping %d bytes)\n", subID, subSize)
					iter.Bytes(subSize)
				}
			}
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
	return instance.enqueueCall(fnIndex, memory.NewStack(params...))
}

func (instance *Instance) enqueueCall(fnIndex int, stack *memory.Stack[any]) (*execution.CallFrame, error) {
	if instance.callStack.Size()+1 > maxCallStackDepth {
		return nil, fmt.Errorf("call stack stack overflow")
	}

	callFrame, err := instance.createCallFrame(fnIndex, stack)
	if err != nil {
		return nil, fmt.Errorf("failed to create enqueue call frame: %w", err)
	}

	instance.callStack.Push(callFrame)

	return callFrame, nil
}

func (instance *Instance) Tick() error {
	// Keep track of the original/root frame for tail calls
	var rootFrame *execution.CallFrame

	for !instance.callStack.IsEmpty() {
		callFrame := instance.callStack.Top()

		if callFrame.Done() {
			instance.callStack.Pop()

			// forward call results to previous frame if it exists
			prev := instance.callStack.Top()
			if prev == nil && rootFrame != nil {
				// No previous frame but we have a root frame from a tail call chain
				// Copy results to the root frame
				rootFrame.Context.Done = true
				prev = rootFrame
			}

			if prev != nil {
				results := callFrame.Context.Results()
				prev.Context.Stack.Push(results...)
			}

			continue
		}

		if err := callFrame.Call(); err != nil {
			return fmt.Errorf("error executing enqueueCall frame: %w", err)
		}

		if callFrame.Context.FunctionCallRequest >= 0 {
			fnIndex := callFrame.Context.FunctionCallRequest
			callFrame.Context.FunctionCallRequest = -1

			if callFrame.Context.TailCall {
				// For tail calls, remember the root frame if this is the first in the chain
				if rootFrame == nil {
					rootFrame = callFrame
				}
				instance.callStack.Pop()
				callFrame.Context.TailCall = false
			}

			stack := callFrame.Context.Stack

			if _, err := instance.enqueueCall(fnIndex, stack); err != nil {
				return fmt.Errorf("failed to enqueue call function at index %d: %w", fnIndex, err)
			}
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

func (instance *Instance) createCallFrame(index int, stack *memory.Stack[any]) (*execution.CallFrame, error) {
	if instance.module.IsFunction(index) {
		return instance.createLocalCallFrame(index, stack)
	}
	if instance.module.IsImport(index) {
		return instance.createImportCallFrame(index, stack)
	}
	return nil, fmt.Errorf("invalid function index: %d", index)
}

func (instance *Instance) createImportCallFrame(index int, stack *memory.Stack[any]) (*execution.CallFrame, error) {
	extFunc, err := instance.getImportedFunc(index)
	if err != nil {
		return nil, fmt.Errorf("invalid function enqueueCall request: %w", err)
	}

	numParams := extFunc.NumInputs()
	numResults := extFunc.NumOutputs()

	if stack.Size() < numParams {
		return nil, fmt.Errorf("not enough parameters on stack for function at index %d: expected %d, got %d", index, numParams, stack.Size())
	}
	params := stack.Last(numParams)

	if instance.debug.flags&verboseShowFunctionCallsFlag != 0 {
		fmt.Printf("Calling imported function at index %d (0x%x) %s with params: %v\n", index, index, extFunc.String(), params)
	}

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      extFunc,

		Context: execution.Context{
			NumParams:  numParams,
			NumResults: numResults,
			Params:     params,

			Stack: memory.NewStack[any](),

			FunctionCallRequest: -1,
		},
	}, nil
}

func (instance *Instance) createLocalCallFrame(index int, stack *memory.Stack[any]) (*execution.CallFrame, error) {
	fn := instance.module.FunctionAt(index)

	numParams := len(fn.Signature.Params)
	numResults := len(fn.Signature.Results)

	if stack.Size() < numParams {
		return nil, fmt.Errorf("not enough parameters on stack for function at index %d: expected %d, got %d", index, numParams, stack.Size())
	}
	params := stack.Last(numParams)

	if instance.debug.flags&verboseShowFunctionCallsFlag != 0 {
		localIndex := index - len(instance.indexedImportedFunctions)
		name := "<unk>"
		if index < len(instance.debug.functions) {
			name = instance.debug.functions[index]
		}
		fmt.Printf("Calling function at index %d (0x%x) $%s with params: %v\n", localIndex, localIndex, name, params)
	}

	debugEnabled := instance.debug.flags&verboseShowInstructions != 0 // TODO: precompute

	locals := memory.NewStackWithCapacity[any](numParams + len(fn.Locals))
	locals.Push(params...)
	locals.Push(fn.Locals...)

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      fn,

		Context: execution.Context{
			NumParams:  numParams,
			NumResults: numResults,
			Params:     params,

			Stack:          memory.NewStack[any](),
			Locals:         locals,
			Globals:        instance.store.Globals,
			Memories:       instance.store.Memories,
			Tables:         instance.store.Tables,
			FuncSignatures: instance.funcSignatures,

			Body:                binary.NewIterator(fn.Body),
			FunctionCallRequest: -1,
			BlockStack:          memory.NewStack[execution.BlockFrame](),
			Blocks:              fn.Blocks,

			Debug: debugEnabled,
		},
	}, nil
}
