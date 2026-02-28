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

func (sl *SkipList) delete(offset int) {
	update := make([]*Node, maxLevel)
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset < offset {
			current = current.forward[i]
		}
		update[i] = current
	}
	target := current.forward[0]
	if target == nil || target.offset != offset {
		return
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
}

func (sl *SkipList) insert(offset, size int, data []byte) {
	update := make([]*Node, maxLevel)
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

	node := newNode(offset, size, sl.minCapacity, newLevel)
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

	// Collect nodes that overlap with [offset, end)
	var toDelete []int
	var leftNode *Node  // partial overlap on the left
	var rightNode *Node // partial overlap on the right

	for n := current.forward[0]; n != nil && n.offset < end; n = n.forward[0] {
		nodeEnd := n.offset + n.size

		if n.offset < offset && nodeEnd > offset {
			// Node extends before our range — save the left portion
			leftNode = &Node{
				offset: n.offset,
				size:   offset - n.offset,
				data:   make([]byte, offset-n.offset),
			}
			copy(leftNode.data, n.data[:offset-n.offset])
		}
		if nodeEnd > end && n.offset < end {
			// Node extends after our range — save the right portion
			rightNode = &Node{
				offset: end,
				size:   nodeEnd - end,
				data:   make([]byte, nodeEnd-end),
			}
			copy(rightNode.data, n.data[end-n.offset:])
		}
		toDelete = append(toDelete, n.offset)
	}

	// Delete all overlapping nodes
	for _, off := range toDelete {
		sl.delete(off)
	}

	// Re-insert trimmed left portion
	if leftNode != nil {
		sl.insert(leftNode.offset, leftNode.size, leftNode.data)
	}

	// Insert the new data
	sl.insert(offset, size, data)

	// Re-insert trimmed right portion
	if rightNode != nil {
		sl.insert(rightNode.offset, rightNode.size, rightNode.data)
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
