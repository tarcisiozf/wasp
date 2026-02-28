package foo

import (
	"fmt"
	"iter"
	"math"
	"math/rand"

	iface "github.com/tarcisiozf/wasp/memory"
)

const maxLevel = 16
const probability = 0.5
const pageSize = 65536 // 64KiB

type Node struct {
	offset   int
	size     int
	capacity int
	data     []byte
	forward  []*Node
}

type SkipList struct {
	head        *Node
	level       int
	length      int
	minCapacity int
	numPages    int
	maxPages    int
	freeNodes   []*Node
}

func (sl *SkipList) Grow(delta int) bool {
	if delta < 0 {
		return false
	}
	if sl.maxPages > 0 && sl.numPages+delta > sl.maxPages {
		return false
	}
	sl.numPages += delta
	return true
}

func (sl *SkipList) NumPages() int {
	return sl.numPages
}

func (sl *SkipList) PageSize() int {
	return pageSize
}

func (sl *SkipList) MaxPages() int {
	return sl.maxPages
}

func (sl *SkipList) Data() []byte {
	//TODO implement me
	panic("implement me")
}

func (sl *SkipList) Clone() iface.Memory {
	//TODO implement me
	panic("implement me")
}

func New(numPages, maxPages, minCapacity int) *SkipList {
	return &SkipList{
		head:        newNode(0, 0, minCapacity, maxLevel),
		level:       1,
		minCapacity: minCapacity,
		numPages:    numPages,
		maxPages:    maxPages,
	}
}

func newNode(offset, size, capacity, level int) *Node {
	capacity = max(size, capacity)
	return &Node{
		offset:   offset,
		size:     size,
		capacity: capacity,
		data:     make([]byte, capacity),
		forward:  make([]*Node, level),
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < probability {
		level++
	}
	return level
}

func (sl *SkipList) delete(offset int) *Node {
	var update [maxLevel]*Node
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset < offset {
			current = current.forward[i]
		}
		update[i] = current
	}
	target := current.forward[0]
	if target == nil || target.offset != offset {
		return nil
	}
	for i := 0; i < sl.level; i++ {
		if update[i].forward[i] != target {
			break
		}
		update[i].forward[i] = target.forward[i]
	}
	sl.length--
	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}
	// Clear forward pointers for reuse
	for i := range target.forward {
		target.forward[i] = nil
	}
	return target
}

func (sl *SkipList) recycleNode(n *Node) {
	sl.freeNodes = append(sl.freeNodes, n)
}

func (sl *SkipList) acquireNode(offset, size, level int) *Node {
	capacity := max(size, sl.minCapacity)
	// Try to find a node from the free list with enough capacity
	if n := len(sl.freeNodes); n > 0 {
		node := sl.freeNodes[n-1]
		sl.freeNodes = sl.freeNodes[:n-1]
		node.offset = offset
		node.size = size
		// Reuse data buffer if large enough
		if cap(node.data) >= capacity {
			node.data = node.data[:capacity]
			node.capacity = capacity
		} else {
			node.data = make([]byte, capacity)
			node.capacity = capacity
		}
		// Reuse forward slice if large enough, otherwise allocate
		if cap(node.forward) >= level {
			node.forward = node.forward[:level]
			for i := range node.forward {
				node.forward[i] = nil
			}
		} else {
			node.forward = make([]*Node, level)
		}
		return node
	}
	return newNode(offset, size, sl.minCapacity, level)
}

func (sl *SkipList) insert(offset, size int, data []byte) {
	var update [maxLevel]*Node
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset < offset {
			current = current.forward[i]
		}
		update[i] = current
	}

	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	node := sl.acquireNode(offset, size, newLevel)
	copy(node.data, data)
	for i := 0; i < newLevel; i++ {
		node.forward[i] = update[i].forward[i]
		update[i].forward[i] = node
	}
	sl.length++
}

func (sl *SkipList) Store(offset int, data []byte) {
	size := len(data)
	if size == 0 {
		return
	}
	end := offset + size

	// Find the node just before the affected range
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset+current.forward[i].size <= offset {
			current = current.forward[i]
		}
	}

	// Fast path: write fits entirely within an existing node
	if n := current.forward[0]; n != nil && n.offset <= offset && n.offset+n.size >= end {
		copy(n.data[offset-n.offset:], data)
		return
	}

	// Stack buffers for left/right trimmed data (avoids heap allocation for typical cases)
	var leftBuf, rightBuf [512]byte
	var leftOffset, leftSize int
	var leftData []byte
	var rightOffset, rightSize int
	var rightData []byte

	// Use a small stack buffer for offsets to delete
	var deleteBuf [32]int
	toDelete := deleteBuf[:0]

	for n := current.forward[0]; n != nil && n.offset < end; n = n.forward[0] {
		nodeEnd := n.offset + n.size

		if n.offset < offset && nodeEnd > offset {
			// Node extends before our range — copy the left portion
			leftOffset = n.offset
			leftSize = offset - n.offset
			if leftSize <= len(leftBuf) {
				leftData = leftBuf[:leftSize]
			} else {
				leftData = make([]byte, leftSize)
			}
			copy(leftData, n.data[:leftSize])
		}
		if nodeEnd > end && n.offset < end {
			// Node extends after our range — copy the right portion
			rightOffset = end
			rightSize = nodeEnd - end
			if rightSize <= len(rightBuf) {
				rightData = rightBuf[:rightSize]
			} else {
				rightData = make([]byte, rightSize)
			}
			copy(rightData, n.data[end-n.offset:n.size])
		}
		toDelete = append(toDelete, n.offset)
	}

	// Delete all overlapping nodes (recycling them)
	for _, off := range toDelete {
		if node := sl.delete(off); node != nil {
			sl.recycleNode(node)
		}
	}

	// Re-insert trimmed left portion
	if leftSize > 0 {
		sl.insert(leftOffset, leftSize, leftData)
	}

	// Insert the new data
	sl.insert(offset, size, data)

	// Re-insert trimmed right portion
	if rightSize > 0 {
		sl.insert(rightOffset, rightSize, rightData)
	}
}

func (sl *SkipList) Load(offset, size int) []byte {
	result := make([]byte, size)
	end := offset + size

	// Find the first node that could intersect [offset, end)
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset+current.forward[i].size <= offset {
			current = current.forward[i]
		}
	}

	// Walk level 0 and copy intersecting chunks
	for current = current.forward[0]; current != nil && current.offset < end; current = current.forward[0] {
		nodeEnd := current.offset + current.size

		// intersection: [max(offset, node.offset), min(end, nodeEnd))
		srcStart := max(offset, current.offset) - current.offset
		srcEnd := min(end, nodeEnd) - current.offset
		dstStart := max(offset, current.offset) - offset

		copy(result[dstStart:], current.data[srcStart:srcEnd])
	}

	return result
}

// Len returns the number of elements.
func (sl *SkipList) Len() int {
	return sl.length
}

// Print displays the skip list structure (useful for debugging).
func (sl *SkipList) Print() {
	fmt.Printf("Skip List (levels: %d, length: %d)\n", sl.level, sl.length)
	for i := sl.level - 1; i >= 0; i-- {
		fmt.Printf("Level %d: head", i)
		node := sl.head.forward[i]
		for node != nil {
			fmt.Printf(" -> %v", node.offset)
			node = node.forward[i]
		}
		fmt.Println(" -> nil")
	}
}

// ExpectedLevels returns the theoretically expected number of levels
// for n elements (useful for capacity planning).
func ExpectedLevels(n int) int {
	if n <= 0 {
		return 1
	}
	return int(math.Log2(float64(n))) + 1
}

func FillMyGap(sl *SkipList, data []byte) {
	for start, end := range chunkify(sl.minCapacity, data) {
		sl.Store(start, data[start:end])
	}
}

func chunkify(chunkThreshold int, data []byte) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		size := len(data)
		start := -1

		for i := 0; i < size; i++ {
			if data[i] == 0 {
				if start >= 0 && i-start >= chunkThreshold {
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
