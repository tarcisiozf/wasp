package module

import (
	"fmt"
	"wasp/wasp/internal/memory"
)

type Module struct {
	functionSignatures []FunctionSignature
	functions          []Function

	exports map[string]Export
	imports map[string]map[string]Import

	startFuncIndex int
}

func NewModule() *Module {
	return &Module{
		exports:        make(map[string]Export),
		imports:        make(map[string]map[string]Import),
		startFuncIndex: -1,
	}
}

func (module *Module) GetExportedFunction(name string) (func(args ...any) ([]any, error), error) {
	export, ok := module.exports[name]
	if !ok {
		return nil, fmt.Errorf("export not found: %s", name)
	}

	if export.kind != kindFunc {
		return nil, fmt.Errorf("export is not a function: %s", name)
	}

	fn := module.FunctionAt(export.index)

	return wrapCallable(fn), nil
}

func (module *Module) StartFunction() (func(args ...any) ([]any, error), error) {
	if module.startFuncIndex < 0 {
		return nil, fmt.Errorf("invalid start function index: %d", module.startFuncIndex)
	}

	fn := module.FunctionAt(module.startFuncIndex)

	return wrapCallable(fn), nil
}

func (module *Module) FunctionAt(index int) Function {
	// function index is offset by number of imports
	return module.functions[index-len(module.imports)]
}

func wrapCallable(fn Function) func(args ...any) ([]any, error) {
	return func(args ...any) ([]any, error) {
		if len(args) != len(fn.signature.params) {
			return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.signature.params), len(args))
		}

		stack := memory.NewStack()

		return fn.call(stack, args)
	}
}
