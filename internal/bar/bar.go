package bar

import (
	"iter"
)

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

type Foo struct {
	root           *Segment
	chunkThreshold int
}

func NewFoo(chunkThreshold int) *Foo {
	return &Foo{
		chunkThreshold: chunkThreshold,
	}
}

func (foo *Foo) Store(offset int, data []byte) {
	for start, end := range foo.chunkify(data) {
		chunkOffset := offset + start
		//chunkSize := end - start

		segments := foo.segmentsForRange(chunkOffset)

		if len(segments) == 0 {
			foo.insertSegment(chunkOffset, data[start:end])
			continue
		}

		seg := foo.mergeSegments(segments)
		seg.set(chunkOffset, data[start:end])
	}
}

func (foo *Foo) chunkify(data []byte) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		size := len(data)
		start := -1

		for i := 0; i < size; i++ {
			if data[i] == 0 {
				if start >= 0 && i-start >= foo.chunkThreshold {
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

func (foo *Foo) segmentsForRange(offset int) (segments []*Segment) {
	if foo.root == nil {
		return nil
	}

	current := foo.root
	for current != nil {
		if offset < current.offset {
			current = current.left
		} else if offset >= current.end {
			current = current.right
		} else {
			segments = append(segments, current)
			current = current.right
		}
	}

	return segments
}

func (foo *Foo) insertSegment(offset int, data []byte) {
	size := len(data)
	seg := &Segment{
		offset: offset,
		end:    offset + size,
		size:   size,
		data:   make([]byte, size),
	}
	seg.set(offset, data)
	foo.insertSegmentNode(seg)
}

func (foo *Foo) mergeSegments(segments []*Segment) *Segment {
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
		foo.removeSegment(s)
	}

	merged := &Segment{
		offset: minOffset,
		end:    maxEnd,
		size:   size,
		data:   data,
	}
	foo.insertSegmentNode(merged)
	return merged
}

func (foo *Foo) removeSegment(s *Segment) {
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
		foo.removeSegment(successor)
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
		foo.root = replacement
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
	foo.rebalanceUp(rebalanceStart)
}

func (foo *Foo) insertSegmentNode(seg *Segment) {
	seg.height = 1
	if foo.root == nil {
		foo.root = seg
		return
	}
	current := foo.root
	for current != nil {
		if seg.offset < current.offset {
			if current.left == nil {
				current.left = seg
				seg.parent = current
				foo.rebalanceUp(current)
				return
			}
			current = current.left
		} else {
			if current.right == nil {
				current.right = seg
				seg.parent = current
				foo.rebalanceUp(current)
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
func (foo *Foo) rotateRight(n *Segment) *Segment {
	l := n.left
	n.left = l.right
	if l.right != nil {
		l.right.parent = n
	}
	l.parent = n.parent
	if n.parent == nil {
		foo.root = l
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
func (foo *Foo) rotateLeft(n *Segment) *Segment {
	r := n.right
	n.right = r.left
	if r.left != nil {
		r.left.parent = n
	}
	r.parent = n.parent
	if n.parent == nil {
		foo.root = r
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
func (foo *Foo) rebalance(n *Segment) *Segment {
	updateHeight(n)
	bf := balanceFactor(n)
	if bf > 1 {
		// left-heavy
		if balanceFactor(n.left) < 0 {
			foo.rotateLeft(n.left) // left-right case
		}
		return foo.rotateRight(n)
	}
	if bf < -1 {
		// right-heavy
		if balanceFactor(n.right) > 0 {
			foo.rotateRight(n.right) // right-left case
		}
		return foo.rotateLeft(n)
	}
	return n
}

// rebalanceUp walks from n toward the root, rebalancing at each ancestor.
func (foo *Foo) rebalanceUp(n *Segment) {
	for n != nil {
		parent := n.parent
		foo.rebalance(n)
		n = parent
	}
}
