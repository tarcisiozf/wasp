package module

import (
	"fmt"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/memory"
)

type Module struct {
	wasm []byte

	functionSignatures []fnsig.Signature
	functions          []funcs.Function

	customSections map[string][]byte
	exports        map[string]Export
	imports        []Import

	startFuncIndex int

	globals  memory.Global
	tables   []memory.Table
	data     []memory.DataSegment
	memories []*memory.Memory
}

func NewModule(wasm []byte) *Module {
	return &Module{
		wasm: wasm,

		exports:        make(map[string]Export),
		customSections: make(map[string][]byte),

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

func (module *Module) GetStartFunction() (int, error) {
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

// TypeSignatures returns a slice of signatures from the type section,
// indexed by type index (used for call_indirect)
func (module *Module) TypeSignatures() []fnsig.Signature {
	return module.functionSignatures
}

// FuncSignatures returns a slice of signatures indexed by function index
// (imports first, then local functions)
func (module *Module) FunctionSignatures() []fnsig.Signature {
	result := make([]fnsig.Signature, len(module.imports)+len(module.functions))
	for i, imp := range module.imports {
		result[i] = imp.Signature
	}
	for i, fn := range module.functions {
		result[len(module.imports)+i] = fn.Signature
	}
	return result
}

func (module *Module) Wasm() []byte {
	return module.wasm
}

func (module *Module) CustomSections() map[string][]byte {
	return module.customSections
}

func (module *Module) DataSegments() []memory.DataSegment {
	return module.data
}
