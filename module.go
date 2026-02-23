package wasp

import (
	"fmt"
	"os"

	"github.com/tarcisiozf/wasp/internal/module"
)

type Module = module.Module

func NewModule(binary []byte) (*module.Module, error) {
	mod := module.NewModule(binary)
	if err := module.Parse(mod); err != nil {
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
