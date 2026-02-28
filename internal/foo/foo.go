package foo

import (
	"fmt"
	"math"
	"math/rand"
)

const maxLevel = 16
const probability = 0.5
const chunkSize = 256

type Node struct {
	offset   int
	size     int
	capacity int
	data     []byte
	forward  []*Node
}

type SkipList struct {
	head   *Node
	level  int
	length int
}

func newNode(offset, size, level int) *Node {
	capacity := min(size, chunkSize)
	return &Node{
		offset:   offset,
		size:     size,
		capacity: capacity,
		data:     make([]byte, 0, capacity),
		forward:  make([]*Node, level),
	}
}

func New() *SkipList {
	return &SkipList{
		head:  newNode(0, 0, maxLevel),
		level: 1,
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < probability {
		level++
	}
	return level
}

func (sl *SkipList) Store(offset int, data []byte) {
	size := len(data)
	update := make([]*Node, maxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset < offset {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	// Update existing node that starts at the same offset
	if current != nil && current.offset == offset {
		current.size = size
		current.capacity = min(size, chunkSize)
		current.data = make([]byte, size)
		copy(current.data, data)
		return
	}

	// Insert new node
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	node := newNode(offset, size, newLevel)
	node.data = make([]byte, size)
	copy(node.data, data)
	for i := 0; i < newLevel; i++ {
		node.forward[i] = update[i].forward[i]
		update[i].forward[i] = node
	}
	sl.length++
}

func (sl *SkipList) Load(offset, size int) []byte {
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].offset+current.forward[i].size <= offset {
			current = current.forward[i]
		}
	}
	current = current.forward[0]

	// Check if the node covers the requested range [offset, offset+size)
	if current != nil && current.offset <= offset && offset+size <= current.offset+current.size {
		start := offset - current.offset
		return current.data[start : start+size]
	}
	return nil
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
