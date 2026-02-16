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
	imports []Import

	startFuncIndex int

	globals        memory.Global
	tables         []memory.Table
	data           []memory.DataSegment
	customSections []memory.CustomSection
	memories       []*memory.Memory
}

func NewModule() *Module {
	return &Module{
		exports: make(map[string]Export),

		startFuncIndex: -1,
	}
}

func (module *Module) GetExportedFunction(name string) (int, error) {
	export, ok := module.exports[name]
	if !ok {
		return -1, fmt.Errorf("export not found: %s", name)
	}

	if export.kind != exportKindFunc {
		return -1, fmt.Errorf("export is not a function: %s", name)
	}

	return export.index, nil
}

func (module *Module) StartFunction() (int, error) {
	if module.startFuncIndex < 0 {
		return -1, fmt.Errorf("module does not have a start function")
	}

	return module.startFuncIndex, nil
}

func (module *Module) FunctionAt(index int) *funcs.Function {
	// function index is offset by number of imports
	return &module.functions[index-len(module.imports)]
}

func (module *Module) Globals() *memory.Global {
	return &module.globals
}

func (module *Module) Imports() []Import {
	return module.imports
}

func (module *Module) IsImport(index int) bool {
	return index >= 0 && index < len(module.imports)
}

func (module *Module) IsFunction(index int) bool {
	return index >= len(module.imports) && index < len(module.imports)+len(module.functions)
}

func (module *Module) Memories() []*memory.Memory {
	return module.memories
}

func (module *Module) Exports() map[string]Export {
	return module.exports
}

func (module *Module) Tables() []memory.Table {
	return module.tables
}
