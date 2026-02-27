package foo

import iface "github.com/tarcisiozf/wasp/memory"

const pageSize = 65536 // 64 KiB

type Segment struct {
	offset, end int
	size        int
	data        []byte
	left, right *Segment
	owner       bool // copy-on-write
}

type FragmentedMemory struct {
	numPages       int
	maxPages       int
	root           *Segment
	sparse         bool
	zeroThreshold  int
	chunkThreshold int
}

func (memory *FragmentedMemory) Clone() iface.Memory {
	markNonOwner(memory.root)
	clone := &FragmentedMemory{
		numPages:       memory.numPages,
		maxPages:       memory.maxPages,
		sparse:         memory.sparse,
		zeroThreshold:  memory.zeroThreshold,
		chunkThreshold: memory.chunkThreshold,
		root:           shallowCloneTree(memory.root),
	}
	return clone
}

// markNonOwner clears the owner flag on every segment in the tree so that
// both the original and the clone will copy-on-write before mutating.
func markNonOwner(seg *Segment) {
	if seg == nil {
		return
	}
	seg.owner = false
	markNonOwner(seg.left)
	markNonOwner(seg.right)
}

// shallowCloneTree duplicates the BST nodes but shares the underlying data slices.
// All cloned segments start as non-owners.
func shallowCloneTree(seg *Segment) *Segment {
	if seg == nil {
		return nil
	}
	return &Segment{
		offset: seg.offset,
		end:    seg.end,
		size:   seg.size,
		data:   seg.data, // shared — COW will copy on first write
		owner:  false,
		left:   shallowCloneTree(seg.left),
		right:  shallowCloneTree(seg.right),
	}
}

func (memory *FragmentedMemory) Data() []byte {
	return memory.Load(0, memory.numPages*pageSize)
}

var _ iface.Memory = (*FragmentedMemory)(nil)

func NewFragmentedMemory(numPages, maxPages int) *FragmentedMemory {
	return &FragmentedMemory{numPages: numPages, maxPages: maxPages}
}

func NewSparseMemory(numPages, maxPages, zeroThreshold, chunkThreshold int) *FragmentedMemory {
	return &FragmentedMemory{
		numPages:       numPages,
		maxPages:       maxPages,
		sparse:         true,
		zeroThreshold:  zeroThreshold,
		chunkThreshold: chunkThreshold,
	}
}

// sparseChunks splits bytes into sub-slices that are worth storing,
// skipping leading/trailing zeros and splitting on zero runs longer than zeroThreshold.
// If chunkThreshold > 0, any span whose size is below that threshold is glued
// forward into the next span (bridging the gap between them), so we never emit
// a segment too small to be worth its own BST node.
// Returns [start, end) index pairs (relative to the start of bytes).
func sparseChunks(bytes []byte, zeroThreshold, chunkThreshold int) [][2]int {
	var result [][2]int
	n := len(bytes)
	i := 0

	for i < n {
		// skip leading zeros
		for i < n && bytes[i] == 0 {
			i++
		}
		if i == n {
			break
		}

		start := i
		lastNonZero := i
		zeroRun := 0

		for i < n {
			if bytes[i] != 0 {
				zeroRun = 0
				lastNonZero = i + 1
			} else {
				zeroRun++
				if zeroRun > zeroThreshold {
					result = append(result, [2]int{start, lastNonZero})
					i++
					for i < n && bytes[i] == 0 {
						i++
					}
					start = -1
					break
				}
			}
			i++
		}

		if start != -1 {
			result = append(result, [2]int{start, lastNonZero})
		}
	}

	// glue tiny spans forward: if a span's size < chunkThreshold, extend it to
	// cover all bytes up to and including the next span (or to its own end if
	// it is the last one).
	if chunkThreshold > 0 {
		for i := 0; i < len(result)-1; {
			curSize := result[i][1] - result[i][0]
			nextSize := result[i+1][1] - result[i+1][0]
			if curSize < chunkThreshold || nextSize < chunkThreshold {
				// either side is tiny — bridge them into one span
				result[i][1] = result[i+1][1]
				result = append(result[:i+1], result[i+2:]...)
				// stay at i — re-examine the merged span against its new neighbour
			} else {
				i++
			}
		}
	}

	return result
}

func (memory *FragmentedMemory) Store(offset int, bytes []byte) {
	size := len(bytes)
	if offset < 0 || offset+size > memory.numPages*pageSize {
		panic("memory access out of bounds")
	}

	if memory.sparse {
		for _, chunk := range sparseChunks(bytes, memory.zeroThreshold, memory.chunkThreshold) {
			memory.storeRaw(offset+chunk[0], bytes[chunk[0]:chunk[1]])
		}
		return
	}

	memory.storeRaw(offset, bytes)
}

func (memory *FragmentedMemory) storeRaw(offset int, bytes []byte) {
	size := len(bytes)
	if size == 0 {
		return
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
		// copy-on-write: unshare before mutating
		if !seg.owner {
			cp := make([]byte, len(seg.data))
			copy(cp, seg.data)
			seg.data = cp
			seg.owner = true
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
		owner:  true,
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
			// glue: extend last — append gives us a new backing array we own
			last.data = append(last.data, cur.data...)
			last.end = cur.end
			last.size = last.end - last.offset
			last.owner = true
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
			// no segment at this offset — find the next one ahead, if any
			next := memory.findNextSegment(offset)
			if next == nil || next.offset >= offset+size {
				break // nothing more to copy
			}
			// advance to where the next segment starts; the gap stays zero
			skip := next.offset - offset
			size -= skip
			offset = next.offset
			continue
		}
		chunk := min(size, seg.end-offset)
		copy(data[originalSize-size:], seg.data[offset-seg.offset:offset-seg.offset+chunk])
		offset += chunk
		size -= chunk
	}
	return data
}

// findNextSegment returns the segment with the smallest offset > given offset.
func (memory *FragmentedMemory) findNextSegment(offset int) *Segment {
	var best *Segment
	current := memory.root
	for current != nil {
		if current.offset > offset {
			best = current
			current = current.left
		} else {
			current = current.right
		}
	}
	return best
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
