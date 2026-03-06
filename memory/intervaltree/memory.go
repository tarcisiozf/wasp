package intervaltree

import (
	"fmt"
	"iter"
	"unsafe"

	iface "github.com/tarcisiozf/wasp/memory"
)

const pageSize = 65536 // 64KiB

type Interval struct {
	Offset      int
	end         int
	size        int
	height      int
	parent      *Interval
	Left, Right *Interval
	Data        []byte
}

func (s Interval) set(offset int, data []byte) {
	start := offset - s.Offset
	copy(s.Data[start:], data)
}

type Memory struct {
	root           *Interval
	numPages       int
	maxPages       int
	chunkThreshold int
	numNodes       int
	size           int

	minStore int
	maxStore int
	minLoad  int
	maxLoad  int
}

func NewMemory(numPages, maxPages, chunkThreshold int) *Memory {
	return &Memory{
		numPages:       numPages,
		maxPages:       maxPages,
		chunkThreshold: chunkThreshold,

		minStore: 1<<63 - 1,
		maxStore: 0,
		minLoad:  1<<63 - 1,
		maxLoad:  0,
	}
}

func (mem *Memory) Store(offset int, data []byte) {
	if len(data) == 0 {
		return
	}

	mem.minStore = min(mem.minStore, offset)
	mem.maxStore = max(mem.maxStore, offset+len(data))

	// Zero out any existing intervals in the full store range,
	// because chunkify will skip zero bytes, but those zeroes
	// must still overwrite previously stored non-zero data.
	fullEnd := offset + len(data)
	for seg := range mem.intervalsForRange(offset, fullEnd) {
		zeroStart := max(seg.Offset, offset)
		zeroEnd := min(seg.end, fullEnd)
		for i := zeroStart - seg.Offset; i < zeroEnd-seg.Offset; i++ {
			seg.Data[i] = 0
		}
	}

	for start, end := range mem.chunkify(data) {
		chunkOffset := offset + start
		chunkEnd := offset + end

		hasIntervals := false
		// intervals are yielded in ascending offset order
		for seg := range mem.intervalsForRange(chunkOffset, chunkEnd) {
			hasIntervals = true

			if chunkOffset >= chunkEnd {
				break
			}

			if chunkOffset < seg.Offset {
				// gap before this interval: insert new interval for the gap
				gapEnd := min(seg.Offset, chunkEnd)
				gapSize := gapEnd - chunkOffset
				mem.insertInterval(chunkOffset, data[start:start+gapSize])
				chunkOffset += gapSize
				start += gapSize
			}

			if chunkOffset >= chunkEnd {
				break
			}

			// write the overlapping portion into this interval
			if chunkOffset >= seg.Offset && chunkOffset < seg.end {
				writeEnd := min(seg.end, chunkEnd)
				writeSize := writeEnd - chunkOffset
				seg.set(chunkOffset, data[start:start+writeSize])
				chunkOffset += writeSize
				start += writeSize
			}
		}

		if !hasIntervals {
			mem.insertInterval(chunkOffset, data[start:end])
			continue
		}

		// gap after the last interval: insert remaining data
		if chunkOffset < chunkEnd {
			mem.insertInterval(chunkOffset, data[start:end])
		}
	}

	// Merge adjacent/overlapping intervals in the affected range.
	mem.mergeRange(offset, fullEnd)
}

func (mem *Memory) Load(offset int, size int) []byte {
	data := make([]byte, size)

	mem.minLoad = min(mem.minLoad, offset)
	mem.maxLoad = max(mem.maxLoad, offset+size)

	for seg := range mem.intervalsForRange(offset, offset+size) {
		segStart := max(seg.Offset, offset)
		segEnd := min(seg.end, offset+size)
		copy(data[segStart-offset:segEnd-offset], seg.Data[segStart-seg.Offset:segEnd-seg.Offset])
	}

	return data
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

func (mem *Memory) Data() []byte {
	return mem.Load(0, mem.NumPages()*mem.PageSize())
}

func (mem *Memory) Clone() iface.Memory {
	clone := &Memory{
		numPages:       mem.numPages,
		maxPages:       mem.maxPages,
		chunkThreshold: mem.chunkThreshold,
		numNodes:       mem.numNodes,
		size:           mem.size,
		root:           cloneInterval(mem.root, nil),
	}
	return clone
}

func cloneInterval(s *Interval, parent *Interval) *Interval {
	if s == nil {
		return nil
	}
	data := make([]byte, len(s.Data))
	copy(data, s.Data)
	node := &Interval{
		Offset: s.Offset,
		end:    s.end,
		size:   s.size,
		height: s.height,
		parent: parent,
		Data:   data,
	}
	node.Left = cloneInterval(s.Left, node)
	node.Right = cloneInterval(s.Right, node)
	return node
}

func (mem *Memory) chunkify(data []byte) iter.Seq2[int, int] {
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

// intervalsForRange yields all intervals overlapping [offset, end) in ascending offset order.
// Uses a fixed-size array stack (zero heap allocation; depth 64 supports >10^13 nodes in AVL).
func (mem *Memory) intervalsForRange(offset, end int) iter.Seq[*Interval] {
	return func(yield func(*Interval) bool) {
		var stack [64]*Interval
		top := 0
		cur := mem.root

		for cur != nil || top > 0 {
			for cur != nil {
				if offset < cur.end {
					stack[top] = cur
					top++
					cur = cur.Left
				} else {
					cur = cur.Right
				}
			}

			if top == 0 {
				break
			}

			top--
			n := stack[top]

			if offset < n.end && end > n.Offset {
				if !yield(n) {
					return
				}
			}

			if end > n.Offset {
				cur = n.Right
			}
		}
	}
}

func (mem *Memory) insertInterval(offset int, data []byte) {
	size := len(data)
	buf := make([]byte, size)
	copy(buf, data)
	seg := &Interval{
		Offset: offset,
		end:    offset + size,
		size:   size,
		Data:   buf,
	}
	mem.insertIntervalNode(seg)
}

func (mem *Memory) insertIntervalNode(seg *Interval) {
	mem.numNodes++
	mem.size += seg.size
	seg.height = 1
	if mem.root == nil {
		mem.root = seg
		return
	}
	current := mem.root
	for current != nil {
		if seg.Offset < current.Offset {
			if current.Left == nil {
				current.Left = seg
				seg.parent = current
				mem.rebalanceUp(current)
				return
			}
			current = current.Left
		} else {
			if current.Right == nil {
				current.Right = seg
				seg.parent = current
				mem.rebalanceUp(current)
				return
			}
			current = current.Right
		}
	}
}

func height(s *Interval) int {
	if s == nil {
		return 0
	}
	return s.height
}

func updateHeight(s *Interval) {
	lh := height(s.Left)
	rh := height(s.Right)
	if lh > rh {
		s.height = lh + 1
	} else {
		s.height = rh + 1
	}
}

func balanceFactor(s *Interval) int {
	return height(s.Left) - height(s.Right)
}

// rotateRight performs a right rotation around n, updating parent links.
func (mem *Memory) rotateRight(n *Interval) *Interval {
	l := n.Left
	n.Left = l.Right
	if l.Right != nil {
		l.Right.parent = n
	}
	l.parent = n.parent
	if n.parent == nil {
		mem.root = l
	} else if n.parent.Left == n {
		n.parent.Left = l
	} else {
		n.parent.Right = l
	}
	l.Right = n
	n.parent = l
	updateHeight(n)
	updateHeight(l)
	return l
}

// rotateLeft performs a left rotation around n, updating parent links.
func (mem *Memory) rotateLeft(n *Interval) *Interval {
	r := n.Right
	n.Right = r.Left
	if r.Left != nil {
		r.Left.parent = n
	}
	r.parent = n.parent
	if n.parent == nil {
		mem.root = r
	} else if n.parent.Left == n {
		n.parent.Left = r
	} else {
		n.parent.Right = r
	}
	r.Left = n
	n.parent = r
	updateHeight(n)
	updateHeight(r)
	return r
}

// rebalance fixes AVL invariant at n and returns the new subtree root.
func (mem *Memory) rebalance(n *Interval) *Interval {
	updateHeight(n)
	bf := balanceFactor(n)
	if bf > 1 {
		// left-heavy
		if balanceFactor(n.Left) < 0 {
			mem.rotateLeft(n.Left) // left-right case
		}
		return mem.rotateRight(n)
	}
	if bf < -1 {
		// right-heavy
		if balanceFactor(n.Right) > 0 {
			mem.rotateRight(n.Right) // right-left case
		}
		return mem.rotateLeft(n)
	}
	return n
}

// rebalanceUp walks from n toward the root, rebalancing at each ancestor.
func (mem *Memory) rebalanceUp(n *Interval) {
	for n != nil {
		parent := n.parent
		mem.rebalance(n)
		n = parent
	}
}

// deleteNode removes n from the AVL tree, rebalancing as needed.
func (mem *Memory) deleteNode(n *Interval) {
	mem.numNodes--
	mem.size -= n.size

	if n.Left != nil && n.Right != nil {
		// find in-order successor (leftmost node in right subtree)
		succ := n.Right
		for succ.Left != nil {
			succ = succ.Left
		}
		// copy successor's data into n
		n.Offset = succ.Offset
		n.end = succ.end
		n.size = succ.size
		n.Data = succ.Data
		// now delete the successor (which has at most one child)
		n = succ
	}

	// n has at most one child
	var child *Interval
	if n.Left != nil {
		child = n.Left
	} else {
		child = n.Right
	}

	parent := n.parent
	if child != nil {
		child.parent = parent
	}

	if parent == nil {
		mem.root = child
	} else if parent.Left == n {
		parent.Left = child
	} else {
		parent.Right = child
	}

	// clear references for GC
	n.Left = nil
	n.Right = nil
	n.parent = nil

	// rebalance from parent upward
	mem.rebalanceUp(parent)
}

// inOrderSuccessor returns the next interval after n in offset order, or nil.
func (mem *Memory) inOrderSuccessor(n *Interval) *Interval {
	if n.Right != nil {
		cur := n.Right
		for cur.Left != nil {
			cur = cur.Left
		}
		return cur
	}
	cur := n
	p := cur.parent
	for p != nil && cur == p.Right {
		cur = p
		p = p.parent
	}
	return p
}

// mergeRange coalesces adjacent or overlapping intervals touching [offset, end).
// Two intervals A, B (A before B) are merged when A.end >= B.Offset (adjacent or overlapping).
func (mem *Memory) mergeRange(offset, end int) {
	// Expand the search range by 1 on each side so we also catch intervals
	// that are exactly adjacent to the written region.
	searchStart := offset - 1
	if searchStart < 0 {
		searchStart = 0
	}
	searchEnd := end + 1

	for {
		// Collect intervals in range. We must re-collect after each merge
		// because deleteNode may swap node data (in-order successor copy),
		// invalidating previously collected pointers.
		var a, b *Interval
		merged := false
		for seg := range mem.intervalsForRange(searchStart, searchEnd) {
			if a == nil {
				a = seg
				continue
			}
			b = seg
			if a.end >= b.Offset {
				// merge b into a
				mergedEnd := b.end
				if a.end > mergedEnd {
					mergedEnd = a.end
				}
				newSize := mergedEnd - a.Offset
				oldSize := a.size
				buf := make([]byte, newSize)
				copy(buf, a.Data)
				// copy b's data (may partially overlap with a's range)
				bStart := b.Offset - a.Offset
				copy(buf[bStart:bStart+b.size], b.Data)

				a.Data = buf
				a.end = mergedEnd
				a.size = newSize
				// adjust mem.size for the growth of a
				mem.size += newSize - oldSize

				mem.deleteNode(b)
				merged = true
				break // restart scan — tree structure changed
			}
			a = b
		}
		if !merged {
			break
		}
	}
}

// MergeAll performs a full in-order merge pass over the entire tree.
func (mem *Memory) MergeAll() {
	if mem.root == nil {
		return
	}
	// find the minimum offset
	minNode := mem.root
	for minNode.Left != nil {
		minNode = minNode.Left
	}
	// find the maximum end
	maxNode := mem.root
	for maxNode.Right != nil {
		maxNode = maxNode.Right
	}
	mem.mergeRange(minNode.Offset, maxNode.end)
}

func (mem *Memory) ChunkThreshold() int {
	return mem.chunkThreshold
}

func (mem *Memory) Size() int {
	return mem.size
}

func (mem *Memory) SizeOf() uint64 {
	nn := uint64(mem.numNodes)
	si := uint64(unsafe.Sizeof(Interval{}))
	sm := uint64(unsafe.Sizeof(Memory{}))
	ss := uint64(mem.size)
	total := (nn * si) + ss + sm
	fmt.Printf("Memory SizeOf: numNodes=%d, sizeof(Interval)=%d, sizeof(Memory)=%d, data=%d, total=%d\n", nn, si, sm, ss, total)
	fmt.Printf("min store offset: %d, max store offset: %d\n", mem.minStore, mem.maxStore)
	fmt.Printf("min load offset: %d, max load offset: %d\n", mem.minLoad, mem.maxLoad)
	return total
}

// Iterate yields all intervals via in-order traversal (zero heap allocation).
func (mem *Memory) Iterate() iter.Seq[*Interval] {
	return func(yield func(*Interval) bool) {
		var stack [64]*Interval
		top := 0
		cur := mem.root

		for cur != nil || top > 0 {
			for cur != nil {
				stack[top] = cur
				top++
				cur = cur.Left
			}

			top--
			n := stack[top]

			if !yield(n) {
				return
			}

			cur = n.Right
		}
	}
}
