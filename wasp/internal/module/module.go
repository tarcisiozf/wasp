package module

import (
	"fmt"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/memory"
)

type DataSegment struct {
	MemoryIndex int
	Offset      int
	Data        []byte
}

type CustomSection struct {
	Name string
	Data []byte
}

type Table struct {
	ElementType byte
	InitialSize int
	MaxSize     int
}

type Module struct {
	functionSignatures []funcs.Signature
	functions          []funcs.Function

	exports map[string]Export
	imports []Import

	startFuncIndex int

	globals        memory.Global
	tables         []Table
	data           []DataSegment
	customSections []CustomSection
	memories       []*memory.Memory
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

	if export.kind != exportKindFunc {
		return funcs.Function{}, fmt.Errorf("export is not a function: %s", name)
	}

	fn := module.FunctionAt(export.index)

	return fn, nil
}

func (module *Module) StartFunction() (funcs.Function, error) {
	if module.startFuncIndex < 0 {
		return funcs.Function{}, fmt.Errorf("module does not have a start function")
	}

	fn := module.FunctionAt(module.startFuncIndex)

	return fn, nil
}

func (module *Module) FunctionAt(index int) funcs.Function {
	// function index is offset by number of imports
	return module.functions[index-len(module.imports)]
}

func (module *Module) Globals() *memory.Global {
	return module.globals.Clone()
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
