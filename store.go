package wasp

import (
	"github.com/tarcisiozf/wasp/internal/bar"
	"github.com/tarcisiozf/wasp/internal/foo"
	"github.com/tarcisiozf/wasp/internal/memory"
	"github.com/tarcisiozf/wasp/internal/module"
	iface "github.com/tarcisiozf/wasp/memory"
)

type StoreOptions func(*Store, *module.Module)

type Store struct {
	Globals  *memory.Global
	Memories []iface.Memory
	Tables   []*memory.Table
}

func WithSkipList(minSize int) StoreOptions {
	return func(store *Store, module *module.Module) {
		moduleMemories := module.Memories()
		store.Memories = make([]iface.Memory, len(moduleMemories))
		for i, mem := range moduleMemories {
			slmem := foo.New(mem.NumPages(), mem.MaxPages(), minSize)
			foo.FillMyGap(slmem, mem.Data())
			store.Memories[i] = slmem
		}
	}
}

func WithSegmentedMemory(chunkThreshold int) StoreOptions {
	return func(store *Store, module *module.Module) {
		moduleMemories := module.Memories()
		store.Memories = make([]iface.Memory, len(moduleMemories))
		for i, mem := range moduleMemories {
			segMem := bar.NewSegmentedMemory(mem.NumPages(), mem.MaxPages(), chunkThreshold)
			segMem.Store(0, mem.Data())
			store.Memories[i] = segMem
		}
	}
}

func WithMemories(memories []iface.Memory) StoreOptions {
	return func(store *Store, module *module.Module) {
		store.Memories = make([]iface.Memory, len(memories))
		for i, mem := range memories {
			store.Memories[i] = mem.Clone()
		}
	}
}

func NewStore(module *module.Module, opts ...StoreOptions) *Store {
	store := &Store{}
	for _, opt := range opts {
		opt(store, module)
	}

	if store.Globals == nil {
		store.Globals = module.Globals().Clone()
	}

	if store.Memories == nil {
		moduleMemories := module.Memories()
		store.Memories = make([]iface.Memory, len(moduleMemories))
		for i, mem := range moduleMemories {
			store.Memories[i] = mem.Clone()
		}
	}

	if store.Tables == nil {
		tables := module.Tables()
		store.Tables = make([]*memory.Table, len(tables))
		for i, tbl := range tables {
			store.Tables[i] = tbl.Clone()
		}
	}

	// Apply data segments to memory
	for _, segment := range module.DataSegments() {
		mem := store.Memories[segment.MemoryIndex]
		mem.Store(segment.Offset, segment.Data)
	}

	return store
}
