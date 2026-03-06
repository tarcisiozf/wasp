package sparse

import (
	"fmt"
	"slices"
	"unsafe"

	iface "github.com/tarcisiozf/wasp/memory"
	"github.com/tarcisiozf/wasp/utils/byteutils"
)

const pageSize = 65536 // 64KiB

var (
	sizeOfByteSlicePtr = uint64(unsafe.Sizeof(&[]byte{}))
	sizeOfByteSlice    = uint64(unsafe.Sizeof([]byte{}))
)

type MemoryOption func(*Memory)

func WithPageMerging(threshold float64) MemoryOption {
	return func(memory *Memory) {
		if threshold < 0 || threshold > 1 {
			panic("threshold must be between 0 and 1")
		}
		memory.mergeThreshold = threshold
	}
}

type Memory struct {
	numPages int
	maxPages int

	pageSize      int
	pagesWithData int
	pages         []*[]byte

	mergeThreshold float64
	foo            []int
	bar            []bool
}

var _ iface.Memory = (*Memory)(nil)

func NewMemory(numPages, maxPages, pageSize int, opts ...MemoryOption) *Memory {
	if !isPowerOfTwo(pageSize) {
		panic("memory size must be a power of two")
	}
	mem := &Memory{
		numPages: numPages,
		maxPages: maxPages,
		pageSize: pageSize,
	}
	for _, opt := range opts {
		opt(mem)
	}
	return mem
}

func NewMemoryWithData(numPages, maxPages, pageSize int, data []byte, opts ...MemoryOption) *Memory {
	size := len(data)
	mem := NewMemory(numPages, maxPages, pageSize, opts...)
	if size == 0 {
		return mem
	}

	for offset := 0; offset < size; offset += pageSize {
		end := min(offset+pageSize, size)
		page := data[offset:end]
		if byteutils.IsEmpty(page) {
			continue
		}
		mem.Store(offset, page)
	}

	return mem
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

	mem.foo = append(mem.foo, size)
	mem.bar = append(mem.bar, byteutils.IsEmpty(bytes))

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

	if mem.mergeThreshold > 0 && mem.shouldMergePages() {
		mem.mergePages()
	}
}

func (mem *Memory) Grow(delta int) bool {
	if delta < 0 {
		return false
	}
	if mem.maxPages > 0 && mem.numPages+delta > mem.maxPages {
		return false
	}
	mem.numPages += delta
	return true
}

func (mem *Memory) NumPages() int {
	return mem.numPages
}

func (mem *Memory) PageSize() int {
	return pageSize
}

func (mem *Memory) MaxPages() int {
	return mem.maxPages
}

func (mem *Memory) Size() int {
	return mem.pagesWithData * mem.pageSize
}

func (mem *Memory) SizeOf() uint64 {
	mem.stats()

	sm := uint64(unsafe.Sizeof(Memory{}))
	slice := uint64(len(mem.pages)) * sizeOfByteSlicePtr
	pages := uint64(mem.pagesWithData) * sizeOfByteSlice
	data := uint64(mem.pagesWithData) * uint64(mem.pageSize)
	return sm + slice + pages + data
}

func (mem *Memory) Data() []byte {
	return mem.Load(0, len(mem.pages)*mem.pageSize)
}

func (mem *Memory) Clone() *Memory {
	pages := make([]*[]byte, len(mem.pages))
	for i, ptr := range mem.pages {
		if ptr == nil {
			continue
		}
		clone := make([]byte, mem.pageSize)
		copy(clone, *ptr)
		pages[i] = &clone
	}

	return &Memory{
		numPages:      mem.numPages,
		maxPages:      mem.maxPages,
		pageSize:      mem.pageSize,
		pagesWithData: mem.pagesWithData,
		pages:         pages,
	}
}

func (mem *Memory) writeToPage(pageIdx int, pageOff int, bytes []byte) {
	var page []byte
	if mem.pages[pageIdx] == nil {
		page = make([]byte, mem.pageSize)
		mem.pages[pageIdx] = &page
		mem.pagesWithData++
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

func (mem *Memory) mergePages() {
	newPageSize := mem.pageSize * 2
	newPages := make([]*[]byte, (len(mem.pages)+1)/2)
	for i := 0; i < len(mem.pages); i += 2 {
		var mergedPage []byte
		if mem.pages[i] != nil {
			mergedPage = append(mergedPage, *mem.pages[i]...)
		} else {
			mergedPage = make([]byte, mem.pageSize)
		}
		if i+1 < len(mem.pages) && mem.pages[i+1] != nil {
			mergedPage = append(mergedPage, *mem.pages[i+1]...)
		} else {
			mergedPage = append(mergedPage, make([]byte, mem.pageSize)...)
		}
		newPages[i/2] = &mergedPage
	}

	pagesWithData := 0
	for _, page := range newPages {
		if page != nil {
			pagesWithData++
		}
	}

	mem.pageSize = newPageSize
	mem.pagesWithData = pagesWithData
	mem.pages = newPages
}

func (mem *Memory) shouldMergePages() bool {
	currentOverhead := calculateOverhead(len(mem.pages), mem.pagesWithData, mem.pageSize)
	if currentOverhead < mem.mergeThreshold {
		return false
	}

	previewOverhead := calculateOverhead(len(mem.pages)/2, mem.pagesWithData, mem.pageSize*2)
	if previewOverhead >= currentOverhead {
		return false
	}

	fmt.Println(
		"savings",
		calculateOverheadCost(len(mem.pages), mem.pagesWithData)-calculateOverheadCost(len(mem.pages)/2, mem.pagesWithData),
	)

	return true
}

func (mem *Memory) stats() {
	slices.Sort(mem.foo)
	var sum uint64
	for _, x := range mem.foo {
		sum += uint64(x)
	}
	count := len(mem.foo)

	fmt.Println("Count", count)
	fmt.Println("Average", sum/uint64(count))
	fmt.Println("Median", mem.foo[count/2])
	fmt.Println("Max", mem.foo[count-1])
	fmt.Println("Min", mem.foo[0])

	{
		bar := make([]int, 0, count)
		var sum uint64
		for i, x := range mem.foo {
			if mem.bar[i] {
				bar = append(bar, x)
				sum += uint64(x)
			}
		}
		slices.Sort(bar)
		count := len(bar)

		fmt.Println("Empty Count", count)
		fmt.Println("Empty Average", sum/uint64(count))
		fmt.Println("Empty Median", bar[count/2])
		fmt.Println("Empty Max", bar[count-1])
	}
}

func calculateOverhead(numPages, numPagesWithData, pageSize int) float64 {
	cost := calculateOverheadCost(numPages, numPagesWithData)
	size := pageSize * numPagesWithData
	return float64(cost) / float64(size)
}

func calculateOverheadCost(numPages, numPagesWithData int) (cost uint64) {
	cost += uint64(numPages) * sizeOfByteSlicePtr      // slice of page pointers
	cost += uint64(numPagesWithData) * sizeOfByteSlice // slice header for the page
	return cost
}
