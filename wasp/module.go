package wasp

import (
	"fmt"
	"os"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/module"
)

type function = funcs.Function

type Module struct {
	internal *module.Module
}

func (module *Module) GetExportedFunction(name string) (function, error) {
	return module.internal.GetExportedFunction(name)
}

func (module *Module) StartFunction() (function, error) {
	return module.internal.StartFunction()
}

func NewModule(binary []byte) (*Module, error) {
	mod := module.NewModule()
	if err := module.Parse(mod, binary); err != nil {
		return nil, fmt.Errorf("failed to parse wasm module: %w", err)
	}

	return &Module{mod}, nil
}

func NewModuleFromFile(path string) (*Module, error) {
	binary, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}
	return NewModule(binary)
}
