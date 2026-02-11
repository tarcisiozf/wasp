package wasp

import (
	"wasp/wasp/internal/memory"
	"wasp/wasp/internal/module"
)

type Store struct {
	globals *memory.Global
}

func NewStore(module *module.Module) *Store {
	return &Store{
		globals: module.Globals.Clone(),
	}
}
