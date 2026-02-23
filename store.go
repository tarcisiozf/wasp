package wasp

import (
	"github.com/tarcisiozf/wasp/internal/memory"
	"github.com/tarcisiozf/wasp/internal/module"
)

type Store struct {
	Globals  *memory.Global
	Memories []*memory.Memory
	Tables   []*memory.Table
}

func NewStore(module *module.Module) *Store {
	globals := module.Globals().Clone()

	memories := module.Memories()
	for i, mem := range memories {
		memories[i] = mem.Clone()
	}

	// Apply data segments to memory
	for _, segment := range module.DataSegments() {
		mem := memories[segment.MemoryIndex]
		mem.Store(segment.Offset, segment.Data)
	}

	tables := module.Tables()
	storeTables := make([]*memory.Table, len(tables))
	for i, tbl := range tables {
		storeTables[i] = tbl.Clone()
	}

	return &Store{
		Globals:  globals,
		Memories: memories,
		Tables:   storeTables,
	}
}
