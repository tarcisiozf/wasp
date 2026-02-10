package memory

type globalEntry struct {
	Mutable bool
	Value   any
}

type Global struct {
	entries []globalEntry
}

func (global *Global) Push(value any, mutable bool) {
	global.entries = append(global.entries, globalEntry{
		Mutable: mutable,
		Value:   value,
	})
}

func (global *Global) Get(index int) any {
	if index < 0 || index >= len(global.entries) {
		panic("global index out of bounds")
	}
	return global.entries[index].Value
}

func (global *Global) Set(index int, pop any) {
	if index < 0 || index >= len(global.entries) {
		panic("global index out of bounds")
	}
	entry := &global.entries[index]
	if !entry.Mutable {
		panic("cannot set immutable global")
	}
	entry.Value = pop
}

func NewGlobal() *Global {
	return &Global{}
}
