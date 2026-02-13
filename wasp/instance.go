package wasp

import (
	"fmt"
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/external"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/module"
	"wasp/wasp/internal/types"
)

type InstanceOption func(*Instance) error

func WithLinker(linker *Linker) InstanceOption {
	return func(e *Instance) error {
		e.linker = linker
		return nil
	}
}

type Instance struct {
	module  *module.Module
	globals *memory.Global

	linker *Linker

	indexedImportedFunctions []*external.Function
}

func NewInstance(module *Module, options ...InstanceOption) (*Instance, error) {
	instance := &Instance{
		module:  module.internal,
		globals: module.internal.Globals(),
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

func (instance *Instance) Call(fn funcs.Function, args ...any) ([]any, error) {
	if len(args) != len(fn.Signature.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.Signature.Params), len(args))
	}

	stack := memory.NewStack[any]()
	ctx := &execution.Context{
		Stack:   stack,
		Globals: instance.globals,

		Body:                binary.NewIterator(fn.Body),
		FunctionCallRequest: -1,
	}

	localDeclCount := ctx.Body.Varint()

	local := make([]any, 0, len(args)+localDeclCount)

	// TODO: args should not be directly used as local variables (?)
	for _, arg := range args {
		local = append(local, arg)
	}

	for i := 0; i < localDeclCount; i++ {
		localTypeCount := ctx.Body.Varint()
		localType := types.ForCode(ctx.Body.Byte())
		for j := 0; j < localTypeCount; j++ {
			local = append(local, localType.Zero())
		}
	}

	ctx.Local = local

	for !ctx.Done {
		opcode := ctx.Body.Byte()
		ix := instructions.Instruction(opcode)
		fmt.Printf("Executing instruction %s\n", ix.String())
		if err := ix.Handler(ctx); err != nil {
			return nil, fmt.Errorf("failed to execute instruction 0x%x: %w", opcode, err)
		}

		if ctx.FunctionCallRequest >= 0 {
			extFunc, err := instance.getImportedFunc(ctx.FunctionCallRequest)
			if err != nil {
				return nil, fmt.Errorf("invalid function call request: %w", err)
			}
			params := ctx.Stack.PopN(extFunc.NumInputs())
			results, err := extFunc.Call(params)
			if err != nil {
				return nil, fmt.Errorf("failed to call external function %s.%s: %w", extFunc.ModuleName(), extFunc.FieldName(), err)
			}
			for _, result := range results {
				ctx.Stack.Push(result)
			}

			ctx.FunctionCallRequest = -1
		}
	}

	results := make([]any, len(fn.Signature.Results))
	for i := range fn.Signature.Results {
		results[i] = ctx.Stack.Pop()
	}
	return results, nil
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
