package wasp

import (
	"fmt"
	"os"
)

const (
	wasmBinaryMagic   = 0x6d736100
	wasmBinaryVersion = 0x1

	sectionType     = 0x1
	sectionFunction = 0x3
	sectionExport   = 0x7
	sectionCode     = 0xa

	typeFunc = 0x60

	exportKindFunc = 0x00

	guessSize = 0x0
)

type Module struct {
	functions []Function
	exports   map[string]Export
}

func NewModule(binary []byte) (*Module, error) {
	module := &Module{
		exports: make(map[string]Export),
	}
	if err := parseModule(module, binary); err != nil {
		return nil, fmt.Errorf("failed to parse wasm module: %w", err)
	}

	return module, nil
}

func NewModuleFromFile(path string) (*Module, error) {
	binary, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}
	return NewModule(binary)
}

func (module *Module) addFunction(params, results []int) int {
	index := len(module.functions)
	module.functions = append(module.functions, Function{
		params:  params,
		results: results,
	})
	return index
}

func (module *Module) GetExportedFunction(name string) (Function, error) {
	export, ok := module.exports[name]
	if !ok {
		return Function{}, fmt.Errorf("export not found: %s", name)
	}

	if export.kind != exportKindFunc {
		return Function{}, fmt.Errorf("export is not a function: %s", name)
	}

	return module.functions[export.index], nil
}
