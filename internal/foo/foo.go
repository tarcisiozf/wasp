package foo

import (
	"fmt"
	"iter"
	"math"
	"math/rand"

	iface "github.com/tarcisiozf/wasp/memory"
)

const maxLevel = 8
const probability = 0.5
const pageSize = 65536 // 64KiB
const maxFreeNodes = 64
const maxFreeNodeDataCap = 4096

type Node struct {
	offset  int
	size    int
	level   uint8
	data    []byte
	forward [maxLevel]*Node
}

type SkipList struct {
	head     *Node
	level    int
	length   int
	numPages int
	maxPages int

	freeNodes []*Node
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
	totalSize := sl.numPages * pageSize
	return sl.Load(0, totalSize)
}

func (sl *SkipList) Clone() iface.Memory {
	clone := &SkipList{
		head:     &Node{},
		level:    sl.level,
		length:   sl.length,
		numPages: sl.numPages,
		maxPages: sl.maxPages,
	}

	// Map from original node to cloned node for forward pointer reconstruction
	nodeMap := make(map[*Node]*Node, sl.length+1)
	nodeMap[sl.head] = clone.head

	// Clone all nodes at level 0
	for n := sl.head.forward[0]; n != nil; n = n.forward[0] {
		data := make([]byte, n.size)
		copy(data, n.data[:n.size])
		cloned := &Node{
			offset: n.offset,
			size:   n.size,
			level:  n.level,
			data:   data,
		}
		nodeMap[n] = cloned
	}

	// Reconstruct forward pointers at all levels
	for n := sl.head; n != nil; n = n.forward[0] {
		cloned := nodeMap[n]
		lvl := int(n.level)
		if n == sl.head {
			lvl = sl.level
		}
		for i := 0; i < lvl; i++ {
			if n.forward[i] != nil {
				cloned.forward[i] = nodeMap[n.forward[i]]
			}
		}
	}

	return clone
}

func New(numPages, maxPages int) *SkipList {
	return &SkipList{
		head:     &Node{},
		level:    1,
		numPages: numPages,
		maxPages: maxPages,
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
	if len(sl.freeNodes) >= maxFreeNodes {
		return // discard — don't grow the pool unboundedly
	}
	// Release large data buffers to avoid retaining too much memory
	if cap(n.data) > maxFreeNodeDataCap {
		n.data = nil
	}
	sl.freeNodes = append(sl.freeNodes, n)
}

func (sl *SkipList) acquireNode(offset, size, level int) *Node {
	// Try to find a node from the free list
	if n := len(sl.freeNodes); n > 0 {
		node := sl.freeNodes[n-1]
		sl.freeNodes = sl.freeNodes[:n-1]
		node.offset = offset
		node.size = size
		node.level = uint8(level)
		// Reuse data buffer if large enough
		if cap(node.data) >= size {
			node.data = node.data[:size]
		} else {
			node.data = make([]byte, size)
		}
		// Clear forward pointers
		for i := range node.forward {
			node.forward[i] = nil
		}
		return node
	}
	return &Node{
		offset: offset,
		size:   size,
		level:  uint8(level),
		data:   make([]byte, size),
	}
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

	// Collect overlapping/adjacent nodes and compute the merged range
	mergeStart := offset
	mergeEnd := end

	var deleteBuf [32]int
	toDelete := deleteBuf[:0]

	for n := current.forward[0]; n != nil && n.offset <= end; n = n.forward[0] {
		nodeEnd := n.offset + n.size

		if nodeEnd < offset {
			continue
		}

		// Expand the merge range
		if n.offset < mergeStart {
			mergeStart = n.offset
		}
		if nodeEnd > mergeEnd {
			mergeEnd = nodeEnd
		}
		toDelete = append(toDelete, n.offset)
	}

	if len(toDelete) == 0 {
		// No overlapping nodes — just insert
		sl.insert(offset, size, data)
		return
	}

	// Build the merged buffer
	mergeSize := mergeEnd - mergeStart
	merged := make([]byte, mergeSize)

	// First, copy data from all overlapping nodes
	for n := current.forward[0]; n != nil && n.offset < mergeEnd; n = n.forward[0] {
		if n.offset+n.size <= mergeStart {
			continue
		}
		dstStart := n.offset - mergeStart
		copy(merged[dstStart:dstStart+n.size], n.data[:n.size])
	}

	// Overwrite with the new data
	copy(merged[offset-mergeStart:], data)

	// Delete all overlapping nodes (recycling them)
	for _, off := range toDelete {
		if node := sl.delete(off); node != nil {
			sl.recycleNode(node)
		}
	}

	// Insert the single merged node
	sl.insert(mergeStart, mergeSize, merged)
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

// FillMyGap initializes the skip list from a contiguous byte slice,
// chunking non-zero regions and appending directly (O(1) per chunk).
func FillMyGap(sl *SkipList, chunkThreshold int, data []byte) {
	var update [maxLevel]*Node
	for i := range update {
		update[i] = sl.head
	}

	for start, end := range chunkify(chunkThreshold, data) {
		chunkSize := end - start
		newLevel := sl.randomLevel()
		if newLevel > sl.level {
			for i := sl.level; i < newLevel; i++ {
				update[i] = sl.head
			}
			sl.level = newLevel
		}

		node := &Node{
			offset: start,
			size:   chunkSize,
			level:  uint8(newLevel),
			data:   make([]byte, chunkSize),
		}
		copy(node.data, data[start:end])

		for i := 0; i < newLevel; i++ {
			node.forward[i] = update[i].forward[i]
			update[i].forward[i] = node
		}
		for i := 0; i < newLevel; i++ {
			update[i] = node
		}
		sl.length++
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
