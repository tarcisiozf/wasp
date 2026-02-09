package wasp

import (
	"fmt"
	"os"
	"wasp/wasp/internal/module"
)

func NewModule(binary []byte) (*module.Module, error) {
	mod := module.NewModule()
	if err := module.Parse(mod, binary); err != nil {
		return nil, fmt.Errorf("failed to parse wasm module: %w", err)
	}

	return mod, nil
}

func NewModuleFromFile(path string) (*module.Module, error) {
	binary, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wasm file: %w", err)
	}
	return NewModule(binary)
}
