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
	Globals             *memory.Global
	Memories            []iface.Memory
	Tables              []*memory.Table
	dataSegmentsApplied bool
}

func WithSkipList(chunkThreshold int) StoreOptions {
	return func(store *Store, module *module.Module) {
		moduleMemories := module.Memories()
		store.Memories = make([]iface.Memory, len(moduleMemories))
		for i, mem := range moduleMemories {
			slmem := foo.New(mem.NumPages(), mem.MaxPages())
			foo.FillMyGap(slmem, chunkThreshold, mem.Data())

			// Apply data segments now so NewStore doesn't double-apply them
			for _, segment := range module.DataSegments() {
				if segment.MemoryIndex == i {
					slmem.Store(segment.Offset, segment.Data)
				}
			}

			store.Memories[i] = slmem
		}
		store.dataSegmentsApplied = true
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

	// Apply data segments to memory (unless already applied by an option)
	if !store.dataSegmentsApplied {
		for _, segment := range module.DataSegments() {
			mem := store.Memories[segment.MemoryIndex]
			mem.Store(segment.Offset, segment.Data)
		}
	}

	return store
}
