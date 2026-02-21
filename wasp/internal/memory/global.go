package memory

type globalEntry struct {
	Mutable bool
	Value   any
}

type Global struct {
	entries []*globalEntry
}

func NewGlobal() *Global {
	return &Global{}
}

func (global *Global) Push(value any, mutable bool) {
	global.entries = append(global.entries, &globalEntry{
		Mutable: mutable,
		Value:   value,
	})
}

func (global *Global) Get(index int) (any, bool, error) {
	if index < 0 || index >= len(global.entries) {
		return nil, false, ErrIndexOutOfBounds
	}
	entry := global.entries[index]
	return entry.Value, entry.Mutable, nil
}

func (global *Global) Set(index int, pop any) error {
	if index < 0 || index >= len(global.entries) {
		return ErrIndexOutOfBounds
	}
	entry := global.entries[index]
	if !entry.Mutable {
		return ErrCannotSetImmutableGlobal
	}
	entry.Value = pop
	return nil
}

func (global *Global) Clone() *Global {
	entries := make([]*globalEntry, len(global.entries))
	for i, entry := range global.entries {
		entries[i] = &globalEntry{
			Mutable: entry.Mutable,
			Value:   entry.Value,
		}
	}

	return &Global{
		entries: entries,
	}
}

func (global *Global) Size() int {
	return len(global.entries)
}
