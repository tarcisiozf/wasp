package memory

const pageSize = 65536 // 64 KiB

type page [pageSize]byte

type Memory struct {
	pages    []page
	maxPages int
}

func NewMemory(numPages, maxPages int) *Memory {
	return &Memory{
		pages:    make([]page, numPages),
		maxPages: maxPages,
	}
}
func (memory *Memory) Size() int {
	return len(memory.pages)
}

func (memory *Memory) Clone() *Memory {
	pages := make([]page, len(memory.pages))
	for i, p := range memory.pages {
		copy(pages[i][:], p[:])
	}
	return &Memory{
		pages:    pages,
		maxPages: memory.maxPages,
	}
}

func (memory *Memory) Grow(delta int) bool {
	if delta < 0 || len(memory.pages)+delta > memory.maxPages {
		return false
	}
	for i := 0; i < delta; i++ {
		memory.pages = append(memory.pages, page{})
	}
	return true
}

func (memory *Memory) Load(offset int, size int) []byte {
	pageIndex := offset / pageSize
	pageOffset := offset % pageSize

	if pageIndex >= len(memory.pages) {
		panic("memory access out of bounds")
	}

	page := memory.pages[pageIndex]
	return page[pageOffset : pageOffset+size]
}

func (memory *Memory) Store(offset int, bytes []byte) {
	pageIndex := offset / pageSize
	pageOffset := offset % pageSize

	if pageIndex >= len(memory.pages) {
		panic("memory access out of bounds")
	}

	copy(memory.pages[pageIndex][pageOffset:], bytes)
}
