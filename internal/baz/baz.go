package baz

import (
	"iter"

	iface "github.com/tarcisiozf/wasp/memory"
)

const pageSize = 65536 // 64KiB
const order = 32       // max entries per leaf / max children per internal node

// entry holds a single contiguous data segment.
type entry struct {
	offset int
	size   int
	data   []byte
}

func (e *entry) end() int { return e.offset + e.size }

// node is either a leaf or an internal node in the B+ tree.
type node struct {
	leaf     bool
	count    int              // number of keys (entries in leaf, keys in internal)
	keys     [order]int       // keys: entry offsets (leaf) or separator keys (internal)
	entries  [order]entry     // only used in leaf nodes
	children [order + 1]*node // only used in internal nodes
	next     *node            // leaf linked list
}

// BPlusTree is a B+ tree-backed sparse memory implementation.
type BPlusTree struct {
	root     *node
	numPages int
	maxPages int
}

func New(numPages, maxPages int) *BPlusTree {
	return &BPlusTree{
		root:     nil,
		numPages: numPages,
		maxPages: maxPages,
	}
}

func (t *BPlusTree) Grow(delta int) bool {
	if delta < 0 {
		return false
	}
	if t.maxPages > 0 && t.numPages+delta > t.maxPages {
		return false
	}
	t.numPages += delta
	return true
}

func (t *BPlusTree) NumPages() int { return t.numPages }
func (t *BPlusTree) PageSize() int { return pageSize }
func (t *BPlusTree) MaxPages() int { return t.maxPages }

func (t *BPlusTree) Data() []byte {
	return t.Load(0, t.numPages*pageSize)
}

func (t *BPlusTree) Clone() iface.Memory {
	clone := &BPlusTree{
		numPages: t.numPages,
		maxPages: t.maxPages,
		root:     cloneNode(t.root),
	}
	// Rebuild leaf next pointers
	if clone.root != nil {
		linkLeaves(clone.root)
	}
	return clone
}

func cloneNode(n *node) *node {
	if n == nil {
		return nil
	}
	c := &node{
		leaf:  n.leaf,
		count: n.count,
		keys:  n.keys,
	}
	if n.leaf {
		for i := 0; i < n.count; i++ {
			e := &n.entries[i]
			data := make([]byte, e.size)
			copy(data, e.data[:e.size])
			c.entries[i] = entry{offset: e.offset, size: e.size, data: data}
		}
	} else {
		for i := 0; i <= n.count; i++ {
			c.children[i] = cloneNode(n.children[i])
		}
	}
	return c
}

// linkLeaves walks the cloned tree and re-links the leaf next pointers.
func linkLeaves(root *node) {
	var prev *node
	var walk func(n *node)
	walk = func(n *node) {
		if n.leaf {
			if prev != nil {
				prev.next = n
			}
			prev = n
			return
		}
		for i := 0; i <= n.count; i++ {
			if n.children[i] != nil {
				walk(n.children[i])
			}
		}
	}
	walk(root)
}

// ---- Search ----

// findLeaf returns the leaf node where offset would reside.
func (t *BPlusTree) findLeaf(offset int) *node {
	n := t.root
	if n == nil {
		return nil
	}
	for !n.leaf {
		i := n.count
		for j := 0; j < n.count; j++ {
			if offset < n.keys[j] {
				i = j
				break
			}
		}
		n = n.children[i]
	}
	return n
}

// findEntry returns the index of the entry in the leaf whose range contains offset,
// or -1 if no such entry exists.
func findEntryContaining(leaf *node, offset int) int {
	for i := 0; i < leaf.count; i++ {
		e := &leaf.entries[i]
		if offset >= e.offset && offset < e.end() {
			return i
		}
		if e.offset > offset {
			break
		}
	}
	return -1
}

// ---- Store ----

func (t *BPlusTree) Store(offset int, data []byte) {
	size := len(data)
	if size == 0 {
		return
	}
	end := offset + size

	if t.root == nil {
		buf := make([]byte, size)
		copy(buf, data)
		leaf := &node{leaf: true, count: 1}
		leaf.keys[0] = offset
		leaf.entries[0] = entry{offset: offset, size: size, data: buf}
		t.root = leaf
		return
	}

	// Find the leaf where our offset would land
	leaf := t.findLeaf(offset)

	// Fast path: write fits entirely within an existing entry in this leaf
	for i := 0; i < leaf.count; i++ {
		e := &leaf.entries[i]
		if e.offset <= offset && e.end() >= end {
			copy(e.data[offset-e.offset:], data)
			return
		}
		if e.offset > offset {
			break
		}
	}

	// Scan for overlapping/adjacent entries starting from the leaf we found.
	// Fast path for single overlap: expand the entry in-place.
	// Slow path for multiple overlaps: collect, merge, delete, insert.

	// First, find exactly what overlaps by scanning leaves
	type overlapInfo struct {
		leaf  *node
		index int
	}
	var overlapBuf [64]overlapInfo
	overlaps := overlapBuf[:0]
	mergeStart := offset
	mergeEnd := end

	scanLeaf := leaf
	for scanLeaf != nil {
		for i := 0; i < scanLeaf.count; i++ {
			e := &scanLeaf.entries[i]
			eEnd := e.end()
			if eEnd < offset {
				continue
			}
			if e.offset > end {
				goto doneCollect
			}
			if e.offset < mergeStart {
				mergeStart = e.offset
			}
			if eEnd > mergeEnd {
				mergeEnd = eEnd
			}
			overlaps = append(overlaps, overlapInfo{scanLeaf, i})
		}
		scanLeaf = scanLeaf.next
	}
doneCollect:

	if len(overlaps) == 0 {
		buf := make([]byte, size)
		copy(buf, data)
		t.insertEntry(entry{offset: offset, size: size, data: buf})
		return
	}

	if len(overlaps) == 1 {
		// Single overlap — expand the entry in-place (no delete+insert)
		ol := overlaps[0]
		e := &ol.leaf.entries[ol.index]
		if mergeStart == e.offset && mergeEnd == e.end() {
			// Same range — just overwrite (already handled by fast path above,
			// but can happen if the write is at the exact boundaries)
			copy(e.data[offset-e.offset:], data)
			return
		}
		// Need to grow the entry
		mergeSize := mergeEnd - mergeStart
		merged := make([]byte, mergeSize)
		// Copy old data
		copy(merged[e.offset-mergeStart:], e.data[:e.size])
		// Overwrite with new data
		copy(merged[offset-mergeStart:], data)
		// Update in place
		oldOffset := e.offset
		e.offset = mergeStart
		e.size = mergeSize
		e.data = merged
		// Update the key in the leaf
		ol.leaf.keys[ol.index] = mergeStart
		// If offset changed and this was the first key in the leaf, update ancestor separators
		if oldOffset != mergeStart && ol.index == 0 {
			t.updateLeafFirstKey(ol.leaf, oldOffset, mergeStart)
		}
		return
	}

	// Multiple overlaps — collect data, delete all, merge, insert
	type collected struct {
		offset int
		size   int
		data   []byte
	}
	var collBuf [64]collected
	coll := collBuf[:0]
	for _, ol := range overlaps {
		e := &ol.leaf.entries[ol.index]
		coll = append(coll, collected{e.offset, e.size, e.data})
	}

	// Build merged buffer
	mergeSize := mergeEnd - mergeStart
	merged := make([]byte, mergeSize)
	for _, c := range coll {
		dstStart := c.offset - mergeStart
		copy(merged[dstStart:dstStart+c.size], c.data[:c.size])
	}
	copy(merged[offset-mergeStart:], data)

	// Delete all overlapping entries
	for _, c := range coll {
		t.deleteByOffset(c.offset)
	}

	// Insert merged entry
	t.insertEntry(entry{offset: mergeStart, size: mergeSize, data: merged})
}

// ---- Load ----

func (t *BPlusTree) Load(offset int, size int) []byte {
	result := make([]byte, size)
	if t.root == nil {
		return result
	}
	end := offset + size

	leaf := t.findLeaf(offset)
	if leaf == nil {
		leaf = t.firstLeaf()
	}

	for leaf != nil {
		for i := 0; i < leaf.count; i++ {
			e := &leaf.entries[i]
			eEnd := e.end()
			if eEnd <= offset {
				continue
			}
			if e.offset >= end {
				return result
			}
			// Intersection
			srcStart := max(offset, e.offset) - e.offset
			srcEnd := min(end, eEnd) - e.offset
			dstStart := max(offset, e.offset) - offset
			copy(result[dstStart:], e.data[srcStart:srcEnd])
		}
		leaf = leaf.next
	}

	return result
}

// ---- Insert ----

func (t *BPlusTree) insertEntry(e entry) {
	if t.root == nil {
		leaf := &node{leaf: true, count: 1}
		leaf.keys[0] = e.offset
		leaf.entries[0] = e
		t.root = leaf
		return
	}

	// Find the target leaf
	path, idxPath := t.pathToLeaf(e.offset)
	leaf := path[len(path)-1]

	// Insert into leaf in sorted order
	insertIntoLeaf(leaf, e)

	// Split if needed, propagating up
	if leaf.count <= order-1 {
		return
	}

	// Need to split
	t.splitUp(path, idxPath)
}

func insertIntoLeaf(leaf *node, e entry) {
	pos := leaf.count
	for i := 0; i < leaf.count; i++ {
		if e.offset < leaf.keys[i] {
			pos = i
			break
		}
	}
	// Shift right
	for i := leaf.count; i > pos; i-- {
		leaf.keys[i] = leaf.keys[i-1]
		leaf.entries[i] = leaf.entries[i-1]
	}
	leaf.keys[pos] = e.offset
	leaf.entries[pos] = e
	leaf.count++
}

// pathToLeaf returns the path from root to the leaf where offset belongs,
// along with the child index taken at each internal node.
func (t *BPlusTree) pathToLeaf(offset int) ([]*node, []int) {
	var path []*node
	var idxPath []int
	n := t.root
	for !n.leaf {
		path = append(path, n)
		i := n.count
		for j := 0; j < n.count; j++ {
			if offset < n.keys[j] {
				i = j
				break
			}
		}
		idxPath = append(idxPath, i)
		n = n.children[i]
	}
	path = append(path, n)
	return path, idxPath
}

func (t *BPlusTree) splitUp(path []*node, idxPath []int) {
	n := path[len(path)-1]

	if n.leaf {
		newLeaf := t.splitLeaf(n)
		promotedKey := newLeaf.keys[0]

		if len(path) == 1 {
			// Root was the leaf — create new root
			newRoot := &node{leaf: false, count: 1}
			newRoot.keys[0] = promotedKey
			newRoot.children[0] = n
			newRoot.children[1] = newLeaf
			t.root = newRoot
			return
		}

		// Insert promoted key into parent
		parent := path[len(path)-2]
		childIdx := idxPath[len(idxPath)-1]
		insertIntoInternal(parent, promotedKey, childIdx, newLeaf)

		if parent.count <= order-1 {
			return
		}

		// Need to split internal nodes up the path
		t.splitInternalUp(path[:len(path)-1], idxPath[:len(idxPath)-1])
	}
}

func (t *BPlusTree) splitInternalUp(path []*node, idxPath []int) {
	for i := len(path) - 1; i >= 0; i-- {
		n := path[i]
		if n.count <= order-1 {
			return
		}
		newInternal, promotedKey := t.splitInternal(n)
		if i == 0 {
			// Root split
			newRoot := &node{leaf: false, count: 1}
			newRoot.keys[0] = promotedKey
			newRoot.children[0] = n
			newRoot.children[1] = newInternal
			t.root = newRoot
			return
		}
		parent := path[i-1]
		childIdx := idxPath[i-1]
		insertIntoInternal(parent, promotedKey, childIdx, newInternal)
	}
}

func insertIntoInternal(parent *node, key int, childIdx int, newChild *node) {
	// Insert key at position childIdx+1, shift right
	pos := childIdx
	for i := parent.count; i > pos; i-- {
		parent.keys[i] = parent.keys[i-1]
	}
	for i := parent.count + 1; i > pos+1; i-- {
		parent.children[i] = parent.children[i-1]
	}
	parent.keys[pos] = key
	parent.children[pos+1] = newChild
	parent.count++
}

func (t *BPlusTree) splitLeaf(leaf *node) *node {
	mid := leaf.count / 2
	newLeaf := &node{leaf: true}
	newLeaf.count = leaf.count - mid

	for i := 0; i < newLeaf.count; i++ {
		newLeaf.keys[i] = leaf.keys[mid+i]
		newLeaf.entries[i] = leaf.entries[mid+i]
	}
	// Clear moved entries from original leaf
	for i := mid; i < leaf.count; i++ {
		leaf.keys[i] = 0
		leaf.entries[i] = entry{}
	}
	leaf.count = mid

	// Maintain linked list
	newLeaf.next = leaf.next
	leaf.next = newLeaf

	return newLeaf
}

func (t *BPlusTree) splitInternal(n *node) (*node, int) {
	mid := n.count / 2
	promotedKey := n.keys[mid]

	newInternal := &node{leaf: false}
	newInternal.count = n.count - mid - 1

	for i := 0; i < newInternal.count; i++ {
		newInternal.keys[i] = n.keys[mid+1+i]
	}
	for i := 0; i <= newInternal.count; i++ {
		newInternal.children[i] = n.children[mid+1+i]
	}

	// Clear moved entries from original
	for i := mid; i < n.count; i++ {
		n.keys[i] = 0
	}
	for i := mid + 1; i <= n.count; i++ {
		n.children[i] = nil
	}
	n.count = mid

	return newInternal, promotedKey
}

// ---- Delete ----

func (t *BPlusTree) deleteByOffset(offset int) {
	if t.root == nil {
		return
	}

	path, idxPath := t.pathToLeaf(offset)
	leaf := path[len(path)-1]

	// Find entry in leaf
	pos := -1
	for i := 0; i < leaf.count; i++ {
		if leaf.keys[i] == offset {
			pos = i
			break
		}
	}
	if pos < 0 {
		return
	}

	// Remove from leaf
	for i := pos; i < leaf.count-1; i++ {
		leaf.keys[i] = leaf.keys[i+1]
		leaf.entries[i] = leaf.entries[i+1]
	}
	leaf.keys[leaf.count-1] = 0
	leaf.entries[leaf.count-1] = entry{}
	leaf.count--

	// If root is a leaf, just check if empty
	if len(path) == 1 {
		if leaf.count == 0 {
			t.root = nil
		}
		return
	}

	// Check underflow: minimum entries = (order-1)/2 for non-root leaves
	minEntries := (order - 1) / 2
	if leaf.count >= minEntries {
		// Update parent key if we deleted the first entry
		if pos == 0 && leaf.count > 0 {
			t.updateAncestorKey(path, idxPath, offset, leaf.keys[0])
		}
		return
	}

	// Underflow — try borrow or merge
	t.handleUnderflow(path, idxPath, offset, pos)
}

func (t *BPlusTree) updateAncestorKey(path []*node, idxPath []int, oldKey, newKey int) {
	// Walk up the path, updating separator keys
	for i := len(path) - 2; i >= 0; i-- {
		parent := path[i]
		childIdx := idxPath[i]
		if childIdx > 0 && parent.keys[childIdx-1] == oldKey {
			parent.keys[childIdx-1] = newKey
			return
		}
	}
}

func (t *BPlusTree) handleUnderflow(path []*node, idxPath []int, deletedKey int, deletedPos int) {
	for level := len(path) - 1; level >= 1; level-- {
		n := path[level]
		parent := path[level-1]
		childIdx := idxPath[level-1]

		minEntries := (order - 1) / 2
		if n.count >= minEntries {
			return
		}

		// Try borrow from left sibling
		if childIdx > 0 {
			leftSib := parent.children[childIdx-1]
			if leftSib.count > minEntries {
				if n.leaf {
					borrowFromLeftLeaf(parent, childIdx, leftSib, n)
				} else {
					borrowFromLeftInternal(parent, childIdx, leftSib, n)
				}
				return
			}
		}

		// Try borrow from right sibling
		if childIdx < parent.count {
			rightSib := parent.children[childIdx+1]
			if rightSib.count > minEntries {
				if n.leaf {
					borrowFromRightLeaf(parent, childIdx, n, rightSib)
				} else {
					borrowFromRightInternal(parent, childIdx, n, rightSib)
				}
				return
			}
		}

		// Merge with a sibling
		if childIdx > 0 {
			leftSib := parent.children[childIdx-1]
			if n.leaf {
				mergeLeaves(parent, childIdx-1, leftSib, n)
			} else {
				mergeInternal(parent, childIdx-1, leftSib, n)
			}
		} else {
			rightSib := parent.children[childIdx+1]
			if n.leaf {
				mergeLeaves(parent, childIdx, n, rightSib)
			} else {
				mergeInternal(parent, childIdx, n, rightSib)
			}
		}

		// Check if root is now empty
		if parent == t.root && parent.count == 0 {
			if parent.children[0] != nil {
				t.root = parent.children[0]
			} else {
				t.root = nil
			}
			return
		}
	}
}

func borrowFromLeftLeaf(parent *node, childIdx int, left, right *node) {
	// Move last entry from left to front of right
	borrowed := left.entries[left.count-1]
	borrowedKey := left.keys[left.count-1]

	// Shift right entries
	for i := right.count; i > 0; i-- {
		right.keys[i] = right.keys[i-1]
		right.entries[i] = right.entries[i-1]
	}
	right.keys[0] = borrowedKey
	right.entries[0] = borrowed
	right.count++

	left.keys[left.count-1] = 0
	left.entries[left.count-1] = entry{}
	left.count--

	parent.keys[childIdx-1] = right.keys[0]
}

func borrowFromRightLeaf(parent *node, childIdx int, left, right *node) {
	// Move first entry from right to end of left
	left.keys[left.count] = right.keys[0]
	left.entries[left.count] = right.entries[0]
	left.count++

	// Shift right entries
	for i := 0; i < right.count-1; i++ {
		right.keys[i] = right.keys[i+1]
		right.entries[i] = right.entries[i+1]
	}
	right.keys[right.count-1] = 0
	right.entries[right.count-1] = entry{}
	right.count--

	parent.keys[childIdx] = right.keys[0]
}

func borrowFromLeftInternal(parent *node, childIdx int, left, right *node) {
	// Shift right's keys and children right
	for i := right.count; i > 0; i-- {
		right.keys[i] = right.keys[i-1]
	}
	for i := right.count + 1; i > 0; i-- {
		right.children[i] = right.children[i-1]
	}
	right.keys[0] = parent.keys[childIdx-1]
	right.children[0] = left.children[left.count]
	right.count++

	parent.keys[childIdx-1] = left.keys[left.count-1]
	left.keys[left.count-1] = 0
	left.children[left.count] = nil
	left.count--
}

func borrowFromRightInternal(parent *node, childIdx int, left, right *node) {
	left.keys[left.count] = parent.keys[childIdx]
	left.children[left.count+1] = right.children[0]
	left.count++

	parent.keys[childIdx] = right.keys[0]

	for i := 0; i < right.count-1; i++ {
		right.keys[i] = right.keys[i+1]
	}
	for i := 0; i < right.count; i++ {
		right.children[i] = right.children[i+1]
	}
	right.keys[right.count-1] = 0
	right.children[right.count] = nil
	right.count--
}

func mergeLeaves(parent *node, sepIdx int, left, right *node) {
	// Merge right into left
	for i := 0; i < right.count; i++ {
		left.keys[left.count+i] = right.keys[i]
		left.entries[left.count+i] = right.entries[i]
	}
	left.count += right.count
	left.next = right.next

	// Remove separator and right child from parent
	for i := sepIdx; i < parent.count-1; i++ {
		parent.keys[i] = parent.keys[i+1]
	}
	for i := sepIdx + 1; i < parent.count; i++ {
		parent.children[i] = parent.children[i+1]
	}
	parent.keys[parent.count-1] = 0
	parent.children[parent.count] = nil
	parent.count--
}

func mergeInternal(parent *node, sepIdx int, left, right *node) {
	// Pull separator down
	left.keys[left.count] = parent.keys[sepIdx]
	left.count++

	// Copy right's keys and children
	for i := 0; i < right.count; i++ {
		left.keys[left.count+i] = right.keys[i]
	}
	for i := 0; i <= right.count; i++ {
		left.children[left.count+i] = right.children[i]
	}
	left.count += right.count

	// Remove separator and right child from parent
	for i := sepIdx; i < parent.count-1; i++ {
		parent.keys[i] = parent.keys[i+1]
	}
	for i := sepIdx + 1; i < parent.count; i++ {
		parent.children[i] = parent.children[i+1]
	}
	parent.keys[parent.count-1] = 0
	parent.children[parent.count] = nil
	parent.count--
}

// updateLeafFirstKey updates ancestor separator keys when a leaf's first key changes.
func (t *BPlusTree) updateLeafFirstKey(leaf *node, oldKey, newKey int) {
if t.root == nil || t.root == leaf {
return
}
n := t.root
for !n.leaf {
for i := 0; i < n.count; i++ {
if n.keys[i] == oldKey {
n.keys[i] = newKey
return
}
}
// Navigate down
i := n.count
for j := 0; j < n.count; j++ {
if oldKey < n.keys[j] {
i = j
break
}
}
n = n.children[i]
}
}


// ---- Helpers ----

func (t *BPlusTree) firstLeaf() *node {
	n := t.root
	if n == nil {
		return nil
	}
	for !n.leaf {
		n = n.children[0]
	}
	return n
}

// Len returns the total number of entries across all leaves.
func (t *BPlusTree) Len() int {
	count := 0
	leaf := t.firstLeaf()
	for leaf != nil {
		count += leaf.count
		leaf = leaf.next
	}
	return count
}

// FillMyGap initializes the B+ tree from a contiguous byte slice,
// chunking non-zero regions and bulk-loading them in sorted order.
func FillMyGap(tree *BPlusTree, chunkThreshold int, data []byte) {
	// Collect all chunks
	type chunk struct {
		offset int
		data   []byte
	}
	var chunks []chunk
	for start, end := range chunkify(chunkThreshold, data) {
		sz := end - start
		buf := make([]byte, sz)
		copy(buf, data[start:end])
		chunks = append(chunks, chunk{offset: start, data: buf})
	}

	if len(chunks) == 0 {
		return
	}

	// Bulk load: build leaves directly, then build internal nodes bottom-up
	// This is much more efficient than inserting one by one.
	entriesPerLeaf := order - 1
	numLeaves := (len(chunks) + entriesPerLeaf - 1) / entriesPerLeaf

	leaves := make([]*node, numLeaves)
	for i := range leaves {
		leaves[i] = &node{leaf: true}
	}

	// Fill leaves
	ci := 0
	for li := 0; li < numLeaves; li++ {
		leaf := leaves[li]
		for leaf.count < entriesPerLeaf && ci < len(chunks) {
			c := chunks[ci]
			leaf.keys[leaf.count] = c.offset
			leaf.entries[leaf.count] = entry{offset: c.offset, size: len(c.data), data: c.data}
			leaf.count++
			ci++
		}
		// Link leaves
		if li > 0 {
			leaves[li-1].next = leaf
		}
	}

	// Build internal nodes bottom-up
	tree.root = buildInternalLevel(leaves)
}

func buildInternalLevel(children []*node) *node {
	if len(children) == 1 {
		return children[0]
	}

	maxKeys := order - 1
	var parents []*node

	i := 0
	for i < len(children) {
		parent := &node{leaf: false}
		parent.children[0] = children[i]
		i++
		for parent.count < maxKeys && i < len(children) {
			// Use first key of the child as separator
			child := children[i]
			var sepKey int
			if child.leaf {
				sepKey = child.keys[0]
			} else {
				// Find leftmost key in subtree
				n := child
				for !n.leaf {
					n = n.children[0]
				}
				sepKey = n.keys[0]
			}
			parent.keys[parent.count] = sepKey
			parent.children[parent.count+1] = child
			parent.count++
			i++
		}
		parents = append(parents, parent)
	}

	return buildInternalLevel(parents)
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
