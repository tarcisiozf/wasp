package bar

import (
	"iter"

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
	for start, end := range mem.chunkify(data) {
		chunkOffset := offset + start
		chunkEnd := offset + end

		segments := mem.segmentsForRange(chunkOffset, chunkEnd)

		if len(segments) == 0 {
			mem.insertSegment(chunkOffset, data[start:end])
			continue
		}

		seg := mem.mergeSegments(segments)
		seg.set(chunkOffset, data[start:end])
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
	//TODO implement me
	panic("implement me")
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

// segmentsForRange returns all segments that overlap the range [offset, end).
func (mem *SegmentedMemory) segmentsForRange(offset, end int) (segments []*Segment) {
	if mem.root == nil {
		return nil
	}

	current := mem.root
	for current != nil {
		if end <= current.offset {
			// range is entirely to the left of this node
			current = current.left
		} else if offset >= current.end {
			// range is entirely to the right of this node
			current = current.right
		} else {
			// overlap: collect and continue right for further overlapping segments
			segments = append(segments, current)
			current = current.right
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

func (mem *SegmentedMemory) mergeSegments(segments []*Segment) *Segment {
	minOffset := segments[0].offset
	maxEnd := segments[0].end
	for _, s := range segments[1:] {
		if s.offset < minOffset {
			minOffset = s.offset
		}
		if s.end > maxEnd {
			maxEnd = s.end
		}
	}

	size := maxEnd - minOffset
	data := make([]byte, size)
	for _, s := range segments {
		start := s.offset - minOffset
		copy(data[start:], s.data)
	}

	for _, s := range segments {
		mem.removeSegment(s)
	}

	merged := &Segment{
		offset: minOffset,
		end:    maxEnd,
		size:   size,
		data:   data,
	}
	mem.insertSegmentNode(merged)
	return merged
}

func (mem *SegmentedMemory) removeSegment(s *Segment) {
	var replacement *Segment

	if s.left == nil {
		replacement = s.right
	} else if s.right == nil {
		replacement = s.left
	} else {
		// find in-order successor (leftmost node in right subtree)
		successor := s.right
		for successor.left != nil {
			successor = successor.left
		}
		// detach successor from its current position
		mem.removeSegment(successor)
		// put successor in place of s
		successor.left = s.left
		successor.right = s.right
		if s.left != nil {
			s.left.parent = successor
		}
		if s.right != nil {
			s.right.parent = successor
		}
		replacement = successor
	}

	if replacement != nil {
		replacement.parent = s.parent
	}
	if s.parent == nil {
		mem.root = replacement
	} else if s.parent.left == s {
		s.parent.left = replacement
	} else {
		s.parent.right = replacement
	}

	// rebalance from the replacement (or the parent if no replacement)
	rebalanceStart := replacement
	if rebalanceStart == nil {
		rebalanceStart = s.parent
	}
	mem.rebalanceUp(rebalanceStart)
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
