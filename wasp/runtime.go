package wasp

import (
	"fmt"
	"wasp/wasp/funcs"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/instructions"
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/module"
	"wasp/wasp/types"
)

type RuntimeOption func(*Runtime) error

func WithLinker(linker *Linker) RuntimeOption {
	return func(e *Runtime) error {
		e.linker = linker
		return nil
	}
}

type Runtime struct {
	linker *Linker
}

func NewRuntime(options ...RuntimeOption) (*Runtime, error) {
	runtime := &Runtime{}
	for _, option := range options {
		if err := option(runtime); err != nil {
			return nil, fmt.Errorf("failed to apply runtime option: %w", err)
		}
	}

	if runtime.linker == nil {
		runtime.linker = NewLinker()
	}

	return runtime, nil
}

func (runtime *Runtime) Call(module *module.Module, fn funcs.Function, args ...any) ([]any, error) {
	if len(args) != len(fn.Signature.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.Signature.Params), len(args))
	}

	stack := memory.NewStack()
	ctx := &execution.Context{
		Stack:               stack,
		Body:                fn.Body,
		FunctionCallRequest: -1,
	}

	localDeclCount := fn.Body.Varint()

	local := make([]any, 0, len(args)+localDeclCount)

	// TODO: args should not be directly used as local variables (?)
	for _, arg := range args {
		local = append(local, arg)
	}

	for i := 0; i < localDeclCount; i++ {
		localTypeCount := fn.Body.Varint()
		localType := fn.Body.Byte()
		for j := 0; j < localTypeCount; j++ {
			local = append(local, zeroValue(localType))
		}
	}

	ctx.Local = local

	for !ctx.Done {
		opcode := fn.Body.Byte()
		ix := instructions.Instruction(opcode)
		if ix.Handler == nil {
			return nil, fmt.Errorf("invalid opcode: 0x%x", opcode)
		}
		if err := ix.Handler(ctx); err != nil {
			return nil, fmt.Errorf("failed to execute instruction 0x%x: %w", opcode, err)
		}

		if ctx.FunctionCallRequest >= 0 {
			imp := module.GetImport(ctx.FunctionCallRequest)

			// TODO: index by name
			extFunc, err := runtime.linker.Get(imp.ModuleName, imp.FieldName)
			if err != nil {
				return nil, fmt.Errorf("failed to find external function %s.%s: %w", imp.ModuleName, imp.FieldName, err)
			}
			if extFunc.NumInputs != len(imp.Signature.Params) {
				return nil, fmt.Errorf("external function %s.%s expects %d parameters, got %d", imp.ModuleName, imp.FieldName, extFunc.NumInputs, len(imp.Signature.Params))
			}
			if extFunc.NumOutputs != len(imp.Signature.Results) {
				return nil, fmt.Errorf("external function %s.%s expects %d results, got %d", imp.ModuleName, imp.FieldName, extFunc.NumOutputs, len(imp.Signature.Results))
			}

			params := ctx.Stack.PopN(extFunc.NumInputs)
			results, err := extFunc.Call(params)
			if err != nil {
				return nil, fmt.Errorf("failed to call external function %s.%s: %w", imp.ModuleName, imp.FieldName, err)
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

func zeroValue(typeCode byte) any {
	switch typeCode {
	case types.Int32:
		return int32(0)
	default:
		panic(fmt.Sprintf("unsupported type code: 0x%x", typeCode))
	}
}
