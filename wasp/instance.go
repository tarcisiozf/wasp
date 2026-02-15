package wasp

import (
	"fmt"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/external"
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
	module   *module.Module
	globals  *memory.Global
	memories []*memory.Memory

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

	callStack := memory.NewStack[*execution.CallFrame]()

	instance := &Instance{
		module:   module,
		globals:  globals,
		memories: memories,

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
	return instance.call(fnIndex, params)
}

func (instance *Instance) call(fnIndex int, params []any) (*execution.CallFrame, error) {
	if instance.callStack.Size()+1 > maxCallStackDepth {
		return nil, fmt.Errorf("call stack overflow")
	}

	callFrame, err := instance.createCallFrame(fnIndex, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create call frame: %w", err)
	}

	instance.callStack.Push(callFrame)

	return callFrame, nil
}

func (instance *Instance) Tick() error {
	for !instance.callStack.IsEmpty() {
		callFrame := instance.callStack.Top()

		//for !ctx.Done {
		//	if ctx.FunctionCallRequest >= 0 {
		//		numParams, err := instance.numParamsForFunc(ctx.FunctionCallRequest)
		//		if err != nil {
		//			return fmt.Errorf("invalid function call request: %w", err)
		//		}
		//		params := ctx.Stack.PopN(numParams)
		//		_, err = instance.call(ctx.FunctionCallRequest, params)
		//		if err != nil {
		//			return fmt.Errorf("failed to call function at index %d: %w", ctx.FunctionCallRequest, err)
		//		}
		//		break fnloop
		//	}

		//if instance.module.IsImport(ctx.FunctionCallRequest) {
		//	extFunc, err := instance.getImportedFunc(ctx.FunctionCallRequest)
		//	if err != nil {
		//		return fmt.Errorf("invalid function call request: %w", err)
		//	}
		//	params := ctx.Stack.PopN(extFunc.NumInputs())
		//	results, err := extFunc.Call(params)
		//	if err != nil {
		//		return fmt.Errorf("failed to call external function %s.%s: %w", extFunc.ModuleName(), extFunc.FieldName(), err)
		//	}
		//	for _, result := range results {
		//		ctx.Stack.Push(result)
		//	}
		//
		//	ctx.FunctionCallRequest = -1
		//} else if instance.module.IsFunction(ctx.FunctionCallRequest) {
		//	foo := instance.module.FunctionAt(ctx.FunctionCallRequest)
		//	params := ctx.Stack.PopN(len(foo.Signature.Params))
		//	_, err := instance.Call(foo, params...)
		//	if err != nil {
		//		return fmt.Errorf("failed to call function at index %d: %w", ctx.FunctionCallRequest, err)
		//	}
		//	for _, result := range results {
		//		ctx.Stack.Push(result)
		//	}
		//
		//	ctx.FunctionCallRequest = -1
		//}

		//
		//}

		if callFrame.Done() {
			instance.callStack.Pop()
			continue
		}

		if err := callFrame.Call(); err != nil {
			return fmt.Errorf("error executing call frame: %w", err)
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

func (instance *Instance) createCallFrame(index int, params []any) (*execution.CallFrame, error) {
	if instance.module.IsFunction(index) {
		return instance.createLocalCallFrame(index, params)
	}
	if instance.module.IsImport(index) {
		return instance.createImportCallFrame(index, params)
	}
	return nil, fmt.Errorf("invalid function index: %d", index)
}

func (instance *Instance) createImportCallFrame(index int, params []any) (*execution.CallFrame, error) {
	extFunc, err := instance.getImportedFunc(index)
	if err != nil {
		return nil, fmt.Errorf("invalid function call request: %w", err)
	}

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      extFunc,
		Context: execution.Context{
			NumParams:  extFunc.NumInputs(),
			NumResults: extFunc.NumOutputs(),
			Params:     params,
		},
	}, nil
}

func (instance *Instance) createLocalCallFrame(index int, params []any) (*execution.CallFrame, error) {
	fn := instance.module.FunctionAt(index)

	stack := memory.NewStack[any]()

	locals := memory.NewStackWithCapacity[any](len(params) + len(fn.Locals))
	locals.Push(params...)
	locals.Push(fn.Locals...)

	return &execution.CallFrame{
		FunctionIndex: index,
		Function:      fn,

		Context: execution.Context{
			NumParams:  len(fn.Signature.Params),
			NumResults: len(fn.Signature.Results),
			Params:     params,

			Stack:    stack,
			Locals:   locals,
			Globals:  instance.globals,
			Memories: instance.memories,

			Body:                binary.NewIterator(fn.Body),
			FunctionCallRequest: -1,
			BlockStack:          memory.NewStack[execution.BlockFrame](),
			Blocks:              fn.Blocks,
		},
	}, nil
}
