package foo

import (
	"fmt"
	"unsafe"

	iface "github.com/tarcisiozf/wasp/memory"
)

const defaultPageSize = 128

type Memory struct {
	numPages int
	maxPages int

	pageSize int
	pages    []*[]byte
}

var _ iface.Memory = (*Memory)(nil)

func NewMemory(numPages, maxPages int) (*Memory, error) {
	return NewMemoryFoo(numPages, maxPages, defaultPageSize)
}

func NewMemoryFoo(numPages, maxPages, pageSize int) (*Memory, error) {
	if !isPowerOfTwo(pageSize) {
		return nil, fmt.Errorf("memory size must be a power of two")
	}
	return &Memory{
		numPages: numPages,
		maxPages: maxPages,
		pageSize: pageSize,
	}, nil
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

func (mem *Memory) Load(offset int, size int) []byte {
	data := make([]byte, size)

	pageIdx := offset / mem.pageSize
	pageOff := offset % mem.pageSize
	written := 0
	for written < size {
		n := min(size-written, mem.pageSize-pageOff)
		mem.loadFromPage(pageIdx, pageOff, data[written:], n)
		written += n
		pageIdx++
		pageOff = 0
	}

	return data
}

func (mem *Memory) Store(offset int, bytes []byte) {
	size := len(bytes)
	if size == 0 {
		return
	}

	mem.ensurePages(offset, size)

	pageIdx := offset / mem.pageSize
	pageOff := offset % mem.pageSize
	written := 0
	for written < size {
		n := min(size-written, mem.pageSize-pageOff)
		mem.writeToPage(pageIdx, pageOff, bytes[written:written+n])
		written += n
		pageIdx++
		pageOff = 0
	}
}

func (mem *Memory) Grow(delta int) bool {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) NumPages() int {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) PageSize() int {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) MaxPages() int {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) Size() int {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) SizeOf() uint64 {
	sm := uint64(unsafe.Sizeof(Memory{}))
	slice := uint64(len(mem.pages)) * uint64(unsafe.Sizeof(&[]byte{}))
	pages := uint64(0)
	data := uint64(0)
	for _, page := range mem.pages {
		if page != nil {
			data += uint64(mem.pageSize)
			pages += uint64(unsafe.Sizeof(*page))
		}
	}
	total := sm + slice + pages + data
	return total
}

func (mem *Memory) Data() []byte {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) Clone() iface.Memory {
	//TODO implement me
	panic("implement me")
}

func (mem *Memory) writeToPage(pageIdx int, pageOff int, bytes []byte) {
	var page []byte
	if mem.pages[pageIdx] == nil {
		page = make([]byte, mem.pageSize)
		mem.pages[pageIdx] = &page
	} else {
		page = *mem.pages[pageIdx]
	}

	copy(page[pageOff:], bytes)
}

func (mem *Memory) ensurePages(offset int, size int) {
	requiredSize := ((offset + size) / mem.pageSize) + 1
	if requiredSize > len(mem.pages) {
		pages := make([]*[]byte, requiredSize)
		copy(pages, mem.pages)
		mem.pages = pages
	}
}

func (mem *Memory) loadFromPage(pageIdx int, pageOff int, dest []byte, n int) {
	if pageIdx >= len(mem.pages) || mem.pages[pageIdx] == nil {
		return
	}
	page := *mem.pages[pageIdx]
	copy(dest, page[pageOff:pageOff+n])
}
