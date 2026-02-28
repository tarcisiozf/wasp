package bar

import (
	"iter"
	"sort"

	iface "github.com/tarcisiozf/wasp/memory"
)

const pageSize = 65536 // 64KiB

type Segment struct {
	offset, end int
	size        int
	height      int
	parent      *Segment
	left, right *Segment
	data        []byte
}

func (s Segment) set(offset int, data []byte) {
	start := offset - s.offset
	copy(s.data[start:], data)
}

type SegmentedMemory struct {
	root           *Segment
	numPages       int
	maxPages       int
	chunkThreshold int
}

func NewSegmentedMemory(numPages, maxPages, chunkThreshold int) *SegmentedMemory {
	return &SegmentedMemory{
		numPages:       numPages,
		maxPages:       maxPages,
		chunkThreshold: chunkThreshold,
	}
}

func (mem *SegmentedMemory) Store(offset int, data []byte) {
	// Zero out any existing segments in the full store range,
	// because chunkify will skip zero bytes, but those zeroes
	// must still overwrite previously stored non-zero data.
	fullEnd := offset + len(data)
	for _, seg := range mem.segmentsForRange(offset, fullEnd) {
		zeroStart := max(seg.offset, offset)
		zeroEnd := min(seg.end, fullEnd)
		for i := zeroStart - seg.offset; i < zeroEnd-seg.offset; i++ {
			seg.data[i] = 0
		}
	}

	for start, end := range mem.chunkify(data) {
		chunkOffset := offset + start
		chunkEnd := offset + end

		segments := mem.segmentsForRange(chunkOffset, chunkEnd)

		if len(segments) == 0 {
			mem.insertSegment(chunkOffset, data[start:end])
			continue
		}

		// sort segments by offset so we process them left-to-right
		sortSegments(segments)

		for _, seg := range segments {
			if chunkOffset >= chunkEnd {
				break
			}

			if chunkOffset < seg.offset {
				// gap before this segment: insert new segment for the gap
				gapEnd := min(seg.offset, chunkEnd)
				gapSize := gapEnd - chunkOffset
				mem.insertSegment(chunkOffset, data[start:start+gapSize])
				chunkOffset += gapSize
				start += gapSize
			}

			if chunkOffset >= chunkEnd {
				break
			}

			// write the overlapping portion into this segment
			if chunkOffset >= seg.offset && chunkOffset < seg.end {
				writeEnd := min(seg.end, chunkEnd)
				writeSize := writeEnd - chunkOffset
				seg.set(chunkOffset, data[start:start+writeSize])
				chunkOffset += writeSize
				start += writeSize
			}
		}

		// gap after the last segment: insert remaining data
		if chunkOffset < chunkEnd {
			mem.insertSegment(chunkOffset, data[start:end])
		}
	}
}

func (mem *SegmentedMemory) Load(offset int, size int) []byte {
	data := make([]byte, size)

	for _, seg := range mem.segmentsForRange(offset, offset+size) {
		segStart := max(seg.offset, offset)
		segEnd := min(seg.end, offset+size)
		copy(data[segStart-offset:segEnd-offset], seg.data[segStart-seg.offset:segEnd-seg.offset])
	}

	return data
}

func (mem *SegmentedMemory) Grow(delta int) bool {
	if delta < 0 {
		return false
	}
	if mem.maxPages > 0 && mem.numPages+delta > mem.maxPages {
		return false
	}
	mem.numPages += delta
	return true
}

func (mem *SegmentedMemory) NumPages() int {
	return mem.numPages
}

func (mem *SegmentedMemory) PageSize() int {
	return pageSize
}

func (mem *SegmentedMemory) MaxPages() int {
	return mem.maxPages
}

func (mem *SegmentedMemory) Data() []byte {
	return mem.Load(0, mem.NumPages()*mem.PageSize())
}

func (mem *SegmentedMemory) Clone() iface.Memory {
	clone := &SegmentedMemory{
		numPages:       mem.numPages,
		maxPages:       mem.maxPages,
		chunkThreshold: mem.chunkThreshold,
		root:           cloneSegment(mem.root, nil),
	}
	return clone
}

func cloneSegment(s *Segment, parent *Segment) *Segment {
	if s == nil {
		return nil
	}
	data := make([]byte, len(s.data))
	copy(data, s.data)
	node := &Segment{
		offset: s.offset,
		end:    s.end,
		size:   s.size,
		height: s.height,
		parent: parent,
		data:   data,
	}
	node.left = cloneSegment(s.left, node)
	node.right = cloneSegment(s.right, node)
	return node
}

func (mem *SegmentedMemory) chunkify(data []byte) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		size := len(data)
		start := -1

		for i := 0; i < size; i++ {
			if data[i] == 0 {
				if start >= 0 && i-start >= mem.chunkThreshold {
					yield(start, i)
					start = -1
				}
			} else if start < 0 {
				start = i
			}
		}

		if start >= 0 {
			yield(start, size)
		}
	}
}

func sortSegments(segments []*Segment) {
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].offset < segments[j].offset
	})
}

// segmentsForRange returns all segments that overlap the range [offset, end).
func (mem *SegmentedMemory) segmentsForRange(offset, end int) (segments []*Segment) {
	if mem.root == nil {
		return nil
	}

	stack := []*Segment{mem.root}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if end <= current.offset {
			// range is entirely to the left of this node
			if current.left != nil {
				stack = append(stack, current.left)
			}
		} else if offset >= current.end {
			// range is entirely to the right of this node
			if current.right != nil {
				stack = append(stack, current.right)
			}
		} else {
			// overlap: collect and search both subtrees
			segments = append(segments, current)
			if current.left != nil {
				stack = append(stack, current.left)
			}
			if current.right != nil {
				stack = append(stack, current.right)
			}
		}
	}

	return segments
}

func (mem *SegmentedMemory) insertSegment(offset int, data []byte) {
	size := len(data)
	seg := &Segment{
		offset: offset,
		end:    offset + size,
		size:   size,
		data:   make([]byte, size),
	}
	seg.set(offset, data)
	mem.insertSegmentNode(seg)
}

func (mem *SegmentedMemory) insertSegmentNode(seg *Segment) {
	seg.height = 1
	if mem.root == nil {
		mem.root = seg
		return
	}
	current := mem.root
	for current != nil {
		if seg.offset < current.offset {
			if current.left == nil {
				current.left = seg
				seg.parent = current
				mem.rebalanceUp(current)
				return
			}
			current = current.left
		} else {
			if current.right == nil {
				current.right = seg
				seg.parent = current
				mem.rebalanceUp(current)
				return
			}
			current = current.right
		}
	}
}

func height(s *Segment) int {
	if s == nil {
		return 0
	}
	return s.height
}

func updateHeight(s *Segment) {
	lh := height(s.left)
	rh := height(s.right)
	if lh > rh {
		s.height = lh + 1
	} else {
		s.height = rh + 1
	}
}

func balanceFactor(s *Segment) int {
	return height(s.left) - height(s.right)
}

// rotateRight performs a right rotation around n, updating parent links.
func (mem *SegmentedMemory) rotateRight(n *Segment) *Segment {
	l := n.left
	n.left = l.right
	if l.right != nil {
		l.right.parent = n
	}
	l.parent = n.parent
	if n.parent == nil {
		mem.root = l
	} else if n.parent.left == n {
		n.parent.left = l
	} else {
		n.parent.right = l
	}
	l.right = n
	n.parent = l
	updateHeight(n)
	updateHeight(l)
	return l
}

// rotateLeft performs a left rotation around n, updating parent links.
func (mem *SegmentedMemory) rotateLeft(n *Segment) *Segment {
	r := n.right
	n.right = r.left
	if r.left != nil {
		r.left.parent = n
	}
	r.parent = n.parent
	if n.parent == nil {
		mem.root = r
	} else if n.parent.left == n {
		n.parent.left = r
	} else {
		n.parent.right = r
	}
	r.left = n
	n.parent = r
	updateHeight(n)
	updateHeight(r)
	return r
}

// rebalance fixes AVL invariant at n and returns the new subtree root.
func (mem *SegmentedMemory) rebalance(n *Segment) *Segment {
	updateHeight(n)
	bf := balanceFactor(n)
	if bf > 1 {
		// left-heavy
		if balanceFactor(n.left) < 0 {
			mem.rotateLeft(n.left) // left-right case
		}
		return mem.rotateRight(n)
	}
	if bf < -1 {
		// right-heavy
		if balanceFactor(n.right) > 0 {
			mem.rotateRight(n.right) // right-left case
		}
		return mem.rotateLeft(n)
	}
	return n
}

// rebalanceUp walks from n toward the root, rebalancing at each ancestor.
func (mem *SegmentedMemory) rebalanceUp(n *Segment) {
	for n != nil {
		parent := n.parent
		mem.rebalance(n)
		n = parent
	}
}
