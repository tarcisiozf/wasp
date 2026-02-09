package module

import (
	"fmt"
	"wasp/wasp/internal/memory"
)

type Module struct {
	functions []Function
	exports   map[string]Export
	imports   map[string]map[string]Import
}

func NewModule() *Module {
	return &Module{
		exports: make(map[string]Export),
		imports: make(map[string]map[string]Import),
	}
}

func (module *Module) addFunction(params, results []int) int {
	index := len(module.functions)
	module.functions = append(module.functions, Function{
		params:  params,
		results: results,
	})
	return index
}

func (module *Module) GetExportedFunction(name string) (func(args ...any) ([]any, error), error) {
	export, ok := module.exports[name]
	if !ok {
		return nil, fmt.Errorf("export not found: %s", name)
	}

	if export.kind != kindFunc {
		return nil, fmt.Errorf("export is not a function: %s", name)
	}

	fn := module.functions[export.index]

	return func(args ...any) ([]any, error) {
		if len(args) != len(fn.params) {
			return nil, fmt.Errorf("expected %d arguments, got %d", len(fn.params), len(args))
		}

		stack := memory.NewStack()

		return fn.call(stack, args)
	}, nil
}
