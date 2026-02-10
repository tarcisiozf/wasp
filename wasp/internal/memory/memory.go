package memory

const pageSize = 65536 // 64 KiB

type page struct {
	Data []byte
}

type Memory struct {
	Pages []*page
}

func NewMemory() *Memory {
	return &Memory{}
}

func (memory *Memory) NumPages() int {
	return len(memory.Pages)
}
