package wasp

import (
	"fmt"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/external"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/module"
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

type Instance struct {
	module         *module.Module
	globals        *memory.Global
	memories       []*memory.Memory
	tables         []memory.Table
	funcSignatures []fnsig.Signature

	linker *Linker

	indexedImportedFunctions []*external.Function

	callStack *memory.Stack[*execution.CallFrame]
}

func NewInstance(module *module.Module, options ...InstanceOption) (*Instance, error) {
	globals := module.Globals().Clone()

	memories := module.Memories()
	for i, mem := range memories {
		memories[i] = mem.Clone()
	}

	tables := module.Tables()
	funcSignatures := module.FunctionSignatures()

	callStack := memory.NewStack[*execution.CallFrame]()

	instance := &Instance{
		module:         module,
		globals:        globals,
		memories:       memories,
		tables:         tables,
		funcSignatures: funcSignatures,

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
	for !instance.callStack.IsEmpty() {
		callFrame := instance.callStack.Top()

		if callFrame.Done() {
			instance.callStack.Pop()

			// forward call results to previous frame if it exists
			prev := instance.callStack.Top()
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
			_, err := instance.enqueueCall(
				callFrame.Context.FunctionCallRequest,
				callFrame.Context.Stack,
			)
			if err != nil {
				return fmt.Errorf("failed to enqueue call function at index %d: %w", callFrame.Context.FunctionCallRequest, err)
			}
			callFrame.Context.FunctionCallRequest = -1
		}

		if callFrame.Context.TailCall {

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
			Globals:        instance.globals,
			Memories:       instance.memories,
			Tables:         instance.tables,
			FuncSignatures: instance.funcSignatures,

			Body:                binary.NewIterator(fn.Body),
			FunctionCallRequest: -1,
			BlockStack:          memory.NewStack[execution.BlockFrame](),
			Blocks:              fn.Blocks,
		},
	}, nil
}
