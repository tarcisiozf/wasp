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
	return &FragmentedMemory{numPages: numPages, maxPages: maxPages}
}

func (memory *FragmentedMemory) Store(offset int, bytes []byte) {
	size := len(bytes)
	if offset < 0 || offset+size > memory.numPages*pageSize {
		panic("memory access out of bounds")
	}

	end := offset + size
	rem := size
	off := offset
	src := bytes

	for rem > 0 {
		seg := memory.findSegment(off)
		if seg == nil {
			break
		}
		chunk := min(rem, seg.end-off)
		copy(seg.data[off-seg.offset:off-seg.offset+chunk], src[:chunk])
		off += chunk
		src = src[chunk:]
		rem -= chunk
	}

	if rem == 0 {
		memory.coalesce()
		return
	}

	newSeg := &Segment{
		offset: off,
		end:    end,
		size:   rem,
		data:   make([]byte, rem),
	}
	copy(newSeg.data, src)
	memory.insertSegment(newSeg)
	memory.coalesce()
}

// coalesce merges all adjacent/touching segments in the BST.
// It rebuilds the tree from a sorted in-order walk.
func (memory *FragmentedMemory) coalesce() {
	// collect sorted segments
	segs := make([]*Segment, 0, 8)
	var collect func(n *Segment)
	collect = func(n *Segment) {
		if n == nil {
			return
		}
		collect(n.left)
		segs = append(segs, n)
		collect(n.right)
	}
	collect(memory.root)

	if len(segs) < 2 {
		return
	}

	// merge adjacent pairs
	merged := segs[:1]
	for i := 1; i < len(segs); i++ {
		last := merged[len(merged)-1]
		cur := segs[i]
		if last.end == cur.offset {
			// glue: extend last
			last.data = append(last.data, cur.data...)
			last.end = cur.end
			last.size = last.end - last.offset
		} else {
			merged = append(merged, cur)
		}
	}

	if len(merged) == len(segs) {
		return // nothing changed
	}

	// rebuild BST from merged list
	for _, s := range merged {
		s.left = nil
		s.right = nil
	}
	memory.root = buildBST(merged)
}

func buildBST(segs []*Segment) *Segment {
	if len(segs) == 0 {
		return nil
	}
	mid := len(segs) / 2
	root := segs[mid]
	root.left = buildBST(segs[:mid])
	root.right = buildBST(segs[mid+1:])
	return root
}

func (memory *FragmentedMemory) Load(offset int, size int) []byte {
	if offset < 0 || offset+size > memory.numPages*pageSize {
		panic("memory access out of bounds")
	}

	originalSize := size
	data := make([]byte, size)
	for size > 0 {
		seg := memory.findSegment(offset)
		if seg == nil {
			break
		}
		chunk := min(size, seg.end-offset)
		copy(data[originalSize-size:], seg.data[offset-seg.offset:offset-seg.offset+chunk])
		offset += chunk
		size -= chunk
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

func (memory *FragmentedMemory) NumPages() int { return memory.numPages }
func (memory *FragmentedMemory) PageSize() int { return pageSize }
func (memory *FragmentedMemory) MaxPages() int { return memory.maxPages }

func (memory *FragmentedMemory) findSegment(offset int) *Segment {
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

func (memory *FragmentedMemory) insertSegment(seg *Segment) {
	if memory.root == nil {
		memory.root = seg
		return
	}
	current := memory.root
	for {
		if seg.offset < current.offset {
			if current.left == nil {
				current.left = seg
				return
			}
			current = current.left
		} else {
			if current.right == nil {
				current.right = seg
				return
			}
			current = current.right
		}
	}
}
