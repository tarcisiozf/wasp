package foo

import iface "github.com/tarcisiozf/wasp/memory"

const pageSize = 65536 // 64 KiB

type Segment struct {
	offset, end int
	size        int
	data        []byte
	left, right *Segment
}

type FragmentedMemory struct {
	numPages int
	maxPages int
	root     *Segment
}

var _ iface.Memory = (*FragmentedMemory)(nil)

func NewFragmentedMemory(numPages, maxPages int) *FragmentedMemory {
	return &FragmentedMemory{
		numPages: numPages,
		maxPages: maxPages,
	}
}

func (memory *FragmentedMemory) Store(offset int, bytes []byte) {
	size := len(bytes)
	if offset < 0 || offset+size > memory.numPages*pageSize {
		panic("memory access out of bounds")
	}

	for size > 0 {
		segment := memory.findSegment(offset)
		if segment == nil {
			segment = &Segment{
				offset: offset,
				end:    offset + size,
				size:   size,
				data:   bytes,
			}
			memory.insertSegment(segment)
			break
		}

		chunkSize := min(size, segment.end-offset)
		copy(segment.data[offset-segment.offset:offset-segment.offset+chunkSize], bytes[len(bytes)-size:len(bytes)-size+chunkSize])
		offset += chunkSize
		size -= chunkSize
	}
}

func (memory *FragmentedMemory) Load(offset int, size int) []byte {
	if offset < 0 || offset+size > memory.numPages*pageSize {
		panic("memory access out of bounds")
	}

	originalSize := size
	data := make([]byte, size)
	for size > 0 {
		segment := memory.findSegment(offset)
		if segment == nil {
			break
		}

		chunkSize := min(size, segment.end-offset)
		copy(data[originalSize-size:], segment.data[offset-segment.offset:offset-segment.offset+chunkSize])
		offset += chunkSize
		size -= chunkSize
	}

	return data
}

func (memory *FragmentedMemory) Grow(delta int) bool {
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

	return true
}

func (memory *FragmentedMemory) NumPages() int {
	return memory.numPages
}

func (memory *FragmentedMemory) PageSize() int {
	return pageSize
}

func (memory *FragmentedMemory) MaxPages() int {
	return memory.maxPages
}

func (memory *FragmentedMemory) findSegment(offset int) *Segment {
	if memory.root == nil {
		return nil
	}

	current := memory.root
	for current != nil {
		if offset < current.offset {
			current = current.left
		} else if offset >= current.end {
			current = current.right
		} else {
			return current
		}
	}

	return nil
}

func (memory *FragmentedMemory) insertSegment(segment *Segment) {
	if memory.root == nil {
		memory.root = segment
		return
	}

	current := memory.root
	for {
		if segment.offset < current.offset {
			if current.left == nil {
				current.left = segment
				return
			}
			current = current.left
		} else {
			if current.right == nil {
				current.right = segment
				return
			}
			current = current.right
		}
	}
}
