package module

import (
	"fmt"
	"wasp/wasp/external"
	"wasp/wasp/funcs"
)

type Module struct {
	functionSignatures []funcs.Signature
	functions          []funcs.Function

	exports map[string]Export
	imports []external.Import

	startFuncIndex int
}

func NewModule() *Module {
	return &Module{
		exports:        make(map[string]Export),
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
	// function index is offset by number of imports
	return module.functions[index-len(module.imports)]
}

func (module *Module) GetImport(index int) external.Import {
	return module.imports[index]
}
