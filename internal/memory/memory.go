package memory

import iface "github.com/tarcisiozf/wasp/memory"

const pageSize = 65536 // 64 KiB

type Memory struct {
	data     []byte
	numPages int
	maxPages int
}

func NewMemory(numPages, maxPages int) *Memory {
	data := make([]byte, numPages*pageSize)
	return &Memory{
		data:     data,
		numPages: numPages,
		maxPages: maxPages,
	}
}

func (memory *Memory) Clone() iface.Memory {
	data := make([]byte, len(memory.data))
	copy(data, memory.data)

	return &Memory{
		data:     data,
		numPages: memory.numPages,
		maxPages: memory.maxPages,
	}
}

func (memory *Memory) Grow(delta int) bool {
	if delta < 0 {
		return false
	}
	if memory.maxPages > 0 && memory.numPages+delta > memory.maxPages {
		return false
	}
	if delta == 0 {
		return true
	}

	memory.numPages += delta
	data := make([]byte, memory.numPages*pageSize)
	copy(data, memory.data)
	memory.data = data

	return true
}

func (memory *Memory) Load(offset int, size int) []byte {
	if offset < 0 || offset+size > len(memory.data) {
		panic("memory access out of bounds")
	}
	return memory.data[offset : offset+size]
}

func (memory *Memory) Store(offset int, bytes []byte) {
	if offset < 0 || offset+len(bytes) > len(memory.data) {
		panic("memory access out of bounds")
	}

	copy(memory.data[offset:], bytes)
}

func (memory *Memory) NumPages() int {
	return memory.numPages
}

func (memory *Memory) PageSize() int {
	return pageSize
}

func (memory *Memory) MaxPages() int {
	return memory.maxPages
}

func (memory *Memory) Data() []byte {
	return memory.data
}
