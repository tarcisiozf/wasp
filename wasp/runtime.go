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

type RuntimeOption func(*Runtime) error

func WithLinker(linker *Linker) RuntimeOption {
	return func(e *Runtime) error {
		e.linker = linker
		return nil
	}
}

type Runtime struct {
	module *module.Module

	linker *Linker

	indexedImportedFunctions []*external.Function
}

func NewRuntime(module *module.Module, options ...RuntimeOption) (*Runtime, error) {
	runtime := &Runtime{
		module: module,
	}
	for _, option := range options {
		if err := option(runtime); err != nil {
			return nil, fmt.Errorf("failed to apply runtime option: %w", err)
		}
	}

	if runtime.linker == nil {
		runtime.linker = NewLinker()
	}

	if err := runtime.mapImportsToExternalFunctions(); err != nil {
		return nil, fmt.Errorf("invalid imports: %w", err)
	}

	return runtime, nil
}

func (runtime *Runtime) Call(fn funcs.Function, args ...any) ([]any, error) {
	if len(args) != len(fn.Signature.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.Signature.Params), len(args))
	}

	stack := memory.NewStack[any]()
	ctx := &execution.Context{
		Stack:   stack,
		Globals: runtime.module.Globals.Clone(),

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
		if ix.Handler == nil {
			return nil, fmt.Errorf("invalid opcode: 0x%x", opcode)
		}
		if err := ix.Handler(ctx); err != nil {
			return nil, fmt.Errorf("failed to execute instruction 0x%x: %w", opcode, err)
		}

		if ctx.FunctionCallRequest >= 0 {
			extFunc := runtime.indexedImportedFunctions[ctx.FunctionCallRequest]
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

func (runtime *Runtime) mapImportsToExternalFunctions() error {
	imports := runtime.module.Imports
	runtime.indexedImportedFunctions = make([]*external.Function, len(imports))

	for i, imp := range imports {
		extFunc, err := runtime.linker.Get(imp.ModuleName, imp.FieldName)
		if err != nil {
			return fmt.Errorf("import %s.%s not found: %w", imp.ModuleName, imp.FieldName, err)
		}
		if err := extFunc.CheckSignatureCompatibility(imp.Signature); err != nil {
			return fmt.Errorf("import %s.%s has incompatible signature: %w", imp.ModuleName, imp.FieldName, err)
		}
		runtime.indexedImportedFunctions[i] = extFunc
	}

	return nil
}
