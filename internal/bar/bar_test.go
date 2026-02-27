package bar

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

func toList(iter iter.Seq2[int, int]) [][2]int {
	var result [][2]int
	iter(func(a, b int) bool {
		result = append(result, [2]int{a, b})
		return true
	})
	return result
}

func TestChunkify(t *testing.T) {
	foo := NewFoo(8)

	t.Run("starts with zero", func(t *testing.T) {
		input := []byte{0, 0, 1, 2, 3, 4, 5, 6, 7, 8}
		expected := [][2]int{{2, 10}}
		result := toList(foo.chunkify(input))
		assert.Equal(t, expected, result)
	})

	t.Run("ends with zero", func(t *testing.T) {
		input := []byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0}
		expected := [][2]int{{0, 8}}
		result := toList(foo.chunkify(input))
		assert.Equal(t, expected, result)
	})

	t.Run("chunk threshold", func(t *testing.T) {
		input := []byte{1, 2, 0, 0, 3, 4, 0, 0, 5, 6, 7, 8}
		expected := [][2]int{{0, 12}}
		result := toList(foo.chunkify(input))
		assert.Equal(t, expected, result)
	})
}

func TestSegmentsForRange(t *testing.T) {
	t.Run("nil root returns nil", func(t *testing.T) {
		foo := NewFoo(8)
		result := foo.segmentsForRange(0)
		assert.Nil(t, result)
	})

	t.Run("offset before segment returns nil", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(10, []byte{1, 2, 3, 4})
		result := foo.segmentsForRange(5)
		assert.Nil(t, result)
	})

	t.Run("offset after segment returns nil", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})
		result := foo.segmentsForRange(10)
		assert.Nil(t, result)
	})

	t.Run("offset within segment returns segment", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})
		result := foo.segmentsForRange(2)
		assert.Len(t, result, 1)
		assert.Equal(t, 0, result[0].offset)
		assert.Equal(t, 4, result[0].end)
	})

	t.Run("offset at segment start is within segment", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(5, []byte{1, 2, 3, 4})
		result := foo.segmentsForRange(5)
		assert.Len(t, result, 1)
		assert.Equal(t, 5, result[0].offset)
	})

	t.Run("offset at segment end is not within segment", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})
		result := foo.segmentsForRange(4)
		assert.Nil(t, result)
	})

	t.Run("returns multiple overlapping segments", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		result := foo.segmentsForRange(4)
		assert.Len(t, result, 2)
		assert.Equal(t, 0, result[0].offset)
		assert.Equal(t, 4, result[1].offset)
	})
}

func TestMergeSegments(t *testing.T) {
	t.Run("single segment is a no-op", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})
		segments := foo.segmentsForRange(2)
		merged := foo.mergeSegments(segments)

		assert.Equal(t, 0, merged.offset)
		assert.Equal(t, 4, merged.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, merged.data)
		assert.Same(t, foo.root, merged)
	})

	t.Run("two overlapping segments are merged into one", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		segments := foo.segmentsForRange(4)
		assert.Len(t, segments, 2)

		merged := foo.mergeSegments(segments)

		assert.Equal(t, 0, merged.offset)
		assert.Equal(t, 8, merged.end)
		assert.Equal(t, 8, merged.size)
	})

	t.Run("earlier segment data is preserved in merged buffer", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		segments := foo.segmentsForRange(4)

		merged := foo.mergeSegments(segments)

		// first segment's data at positions 0-3 should be intact
		assert.Equal(t, byte(1), merged.data[0])
		assert.Equal(t, byte(4), merged.data[3])
	})

	t.Run("later segment data overwrites earlier in the merged buffer", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		segments := foo.segmentsForRange(4)

		merged := foo.mergeSegments(segments)

		// second segment starts at offset 4, so data[4..8] should be from it
		assert.Equal(t, byte(9), merged.data[4])
		assert.Equal(t, byte(12), merged.data[7])
	})

	t.Run("old segments are removed from BST", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		segments := foo.segmentsForRange(4)

		foo.mergeSegments(segments)

		// only one segment should remain in the tree
		assert.Nil(t, foo.root.left)
		assert.Nil(t, foo.root.right)
	})

	t.Run("merged segment is accessible via segmentsForRange", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		foo.insertSegment(4, []byte{9, 10, 11, 12})
		segments := foo.segmentsForRange(4)

		merged := foo.mergeSegments(segments)

		found := foo.segmentsForRange(6)
		assert.Len(t, found, 1)
		assert.Same(t, merged, found[0])
	})
}

func treeHeight(s *Segment) int {
	if s == nil {
		return 0
	}
	l := treeHeight(s.left)
	r := treeHeight(s.right)
	if l > r {
		return l + 1
	}
	return r + 1
}

func isBalanced(s *Segment) bool {
	if s == nil {
		return true
	}
	bf := treeHeight(s.left) - treeHeight(s.right)
	if bf < -1 || bf > 1 {
		return false
	}
	return isBalanced(s.left) && isBalanced(s.right)
}

func TestAVLBalance(t *testing.T) {
	t.Run("right-skewed inserts trigger left rotation", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1})
		foo.insertSegment(10, []byte{1})
		foo.insertSegment(20, []byte{1})
		// without balancing root would be 0 with a right-skewed chain
		assert.Equal(t, 10, foo.root.offset)
		assert.Equal(t, 0, foo.root.left.offset)
		assert.Equal(t, 20, foo.root.right.offset)
		assert.True(t, isBalanced(foo.root))
	})

	t.Run("left-skewed inserts trigger right rotation", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(20, []byte{1})
		foo.insertSegment(10, []byte{1})
		foo.insertSegment(0, []byte{1})
		assert.Equal(t, 10, foo.root.offset)
		assert.Equal(t, 0, foo.root.left.offset)
		assert.Equal(t, 20, foo.root.right.offset)
		assert.True(t, isBalanced(foo.root))
	})

	t.Run("left-right case triggers double rotation", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(20, []byte{1})
		foo.insertSegment(0, []byte{1})
		foo.insertSegment(10, []byte{1})
		assert.Equal(t, 10, foo.root.offset)
		assert.True(t, isBalanced(foo.root))
	})

	t.Run("right-left case triggers double rotation", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1})
		foo.insertSegment(20, []byte{1})
		foo.insertSegment(10, []byte{1})
		assert.Equal(t, 10, foo.root.offset)
		assert.True(t, isBalanced(foo.root))
	})

	t.Run("many sequential inserts stay balanced", func(t *testing.T) {
		foo := NewFoo(8)
		for i := 0; i < 20; i++ {
			foo.insertSegment(i*10, []byte{1})
		}
		assert.True(t, isBalanced(foo.root))
	})

	t.Run("tree stays balanced after removes", func(t *testing.T) {
		foo := NewFoo(8)
		for i := 0; i < 10; i++ {
			foo.insertSegment(i*10, []byte{1})
		}
		// remove every other segment
		for i := 0; i < 10; i += 2 {
			segs := foo.segmentsForRange(i * 10)
			if len(segs) > 0 {
				foo.removeSegment(segs[0])
			}
		}
		assert.True(t, isBalanced(foo.root))
	})
}

func TestInsertSegment(t *testing.T) {
	t.Run("root segment", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})

		assert.NotNil(t, foo.root)
		assert.Equal(t, 0, foo.root.offset)
		assert.Equal(t, 4, foo.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, foo.root.data)
	})

	t.Run("non-overlapping segment", func(t *testing.T) {
		foo := NewFoo(8)
		foo.insertSegment(0, []byte{1, 2, 3, 4})
		foo.insertSegment(10, []byte{5, 6, 7, 8})

		assert.NotNil(t, foo.root)
		assert.Equal(t, 0, foo.root.offset)
		assert.Equal(t, 4, foo.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, foo.root.data)

		assert.NotNil(t, foo.root.right)
		assert.Equal(t, 10, foo.root.right.offset)
		assert.Equal(t, 14, foo.root.right.end)
		assert.Equal(t, []byte{5, 6, 7, 8}, foo.root.right.data)
	})
}
