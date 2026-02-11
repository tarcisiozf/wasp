package module

import (
	"fmt"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/memory"
)

type Module struct {
	functionSignatures []funcs.Signature
	functions          []funcs.Function

	exports map[string]Export
	Imports []Import

	startFuncIndex int

	Globals memory.Global
	Tables  []Table
}

func NewModule() *Module {
	return &Module{
		exports: make(map[string]Export),

		startFuncIndex: -1,
	}
}

func (module *Module) GetExportedFunction(name string) (funcs.Function, error) {
	export, ok := module.exports[name]
	if !ok {
		return funcs.Function{}, fmt.Errorf("export not found: %s", name)
	}

	if export.kind != kindFunc {
		return funcs.Function{}, fmt.Errorf("export is not a function: %s", name)
	}

	fn := module.FunctionAt(export.index)

	return fn, nil
}

func (module *Module) StartFunction() (funcs.Function, error) {
	if module.startFuncIndex < 0 {
		return funcs.Function{}, fmt.Errorf("invalid start function index: %d", module.startFuncIndex)
	}

	fn := module.FunctionAt(module.startFuncIndex)

	return fn, nil
}

func (module *Module) FunctionAt(index int) funcs.Function {
	// function index is offset by number of Imports
	return module.functions[index-len(module.Imports)]
}
