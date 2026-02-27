package bar

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	numPages       = 0
	maxPages       = 10
	chunkThreshold = 8
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
	mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)

	t.Run("starts with zero", func(t *testing.T) {
		input := []byte{0, 0, 1, 2, 3, 4, 5, 6, 7, 8}
		expected := [][2]int{{2, 10}}
		result := toList(mem.chunkify(input))
		assert.Equal(t, expected, result)
	})

	t.Run("ends with zero", func(t *testing.T) {
		input := []byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0}
		expected := [][2]int{{0, 8}}
		result := toList(mem.chunkify(input))
		assert.Equal(t, expected, result)
	})

	t.Run("chunk threshold", func(t *testing.T) {
		input := []byte{1, 2, 0, 0, 3, 4, 0, 0, 5, 6, 7, 8}
		expected := [][2]int{{0, 12}}
		result := toList(mem.chunkify(input))
		assert.Equal(t, expected, result)
	})
}

func TestSegmentsForRange(t *testing.T) {
	t.Run("nil root returns nil", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		result := mem.segmentsForRange(0, 1)
		assert.Nil(t, result)
	})

	t.Run("offset before segment returns nil", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(10, []byte{1, 2, 3, 4})
		result := mem.segmentsForRange(5, 6)
		assert.Nil(t, result)
	})

	t.Run("offset after segment returns nil", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		result := mem.segmentsForRange(10, 11)
		assert.Nil(t, result)
	})

	t.Run("offset within segment returns segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		result := mem.segmentsForRange(2, 3)
		assert.Len(t, result, 1)
		assert.Equal(t, 0, result[0].offset)
		assert.Equal(t, 4, result[0].end)
	})

	t.Run("offset at segment start is within segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(5, []byte{1, 2, 3, 4})
		result := mem.segmentsForRange(5, 6)
		assert.Len(t, result, 1)
		assert.Equal(t, 5, result[0].offset)
	})

	t.Run("offset at segment end is not within segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		result := mem.segmentsForRange(4, 5)
		assert.Nil(t, result)
	})

	t.Run("returns multiple overlapping segments", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		result := mem.segmentsForRange(4, 5)
		assert.Len(t, result, 2)
		assert.Equal(t, 0, result[0].offset)
		assert.Equal(t, 4, result[1].offset)
	})

	t.Run("range spanning two separate segments returns both", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		mem.insertSegment(8, []byte{5, 6, 7, 8})
		result := mem.segmentsForRange(2, 10)
		assert.Len(t, result, 2)
		assert.Equal(t, 0, result[0].offset)
		assert.Equal(t, 8, result[1].offset)
	})
}

func TestMergeSegments(t *testing.T) {
	t.Run("single segment is a no-op", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		segments := mem.segmentsForRange(2, 3)
		merged := mem.mergeSegments(segments)

		assert.Equal(t, 0, merged.offset)
		assert.Equal(t, 4, merged.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, merged.data)
		assert.Same(t, mem.root, merged)
	})

	t.Run("two overlapping segments are merged into one", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		segments := mem.segmentsForRange(4, 5)
		assert.Len(t, segments, 2)

		merged := mem.mergeSegments(segments)

		assert.Equal(t, 0, merged.offset)
		assert.Equal(t, 8, merged.end)
		assert.Equal(t, 8, merged.size)
	})

	t.Run("earlier segment data is preserved in merged buffer", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		segments := mem.segmentsForRange(4, 5)

		merged := mem.mergeSegments(segments)

		// first segment's data at positions 0-3 should be intact
		assert.Equal(t, byte(1), merged.data[0])
		assert.Equal(t, byte(4), merged.data[3])
	})

	t.Run("later segment data overwrites earlier in the merged buffer", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		segments := mem.segmentsForRange(4, 5)

		merged := mem.mergeSegments(segments)

		// second segment starts at offset 4, so data[4..8] should be from it
		assert.Equal(t, byte(9), merged.data[4])
		assert.Equal(t, byte(12), merged.data[7])
	})

	t.Run("old segments are removed from BST", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		segments := mem.segmentsForRange(4, 5)

		mem.mergeSegments(segments)

		// only one segment should remain in the tree
		assert.Nil(t, mem.root.left)
		assert.Nil(t, mem.root.right)
	})

	t.Run("merged segment is accessible via segmentsForRange", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.insertSegment(4, []byte{9, 10, 11, 12})
		segments := mem.segmentsForRange(4, 5)

		merged := mem.mergeSegments(segments)

		found := mem.segmentsForRange(6, 7)
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
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1})
		mem.insertSegment(10, []byte{1})
		mem.insertSegment(20, []byte{1})
		// without balancing root would be 0 with a right-skewed chain
		assert.Equal(t, 10, mem.root.offset)
		assert.Equal(t, 0, mem.root.left.offset)
		assert.Equal(t, 20, mem.root.right.offset)
		assert.True(t, isBalanced(mem.root))
	})

	t.Run("left-skewed inserts trigger right rotation", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(20, []byte{1})
		mem.insertSegment(10, []byte{1})
		mem.insertSegment(0, []byte{1})
		assert.Equal(t, 10, mem.root.offset)
		assert.Equal(t, 0, mem.root.left.offset)
		assert.Equal(t, 20, mem.root.right.offset)
		assert.True(t, isBalanced(mem.root))
	})

	t.Run("left-right case triggers double rotation", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(20, []byte{1})
		mem.insertSegment(0, []byte{1})
		mem.insertSegment(10, []byte{1})
		assert.Equal(t, 10, mem.root.offset)
		assert.True(t, isBalanced(mem.root))
	})

	t.Run("right-left case triggers double rotation", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1})
		mem.insertSegment(20, []byte{1})
		mem.insertSegment(10, []byte{1})
		assert.Equal(t, 10, mem.root.offset)
		assert.True(t, isBalanced(mem.root))
	})

	t.Run("many sequential inserts stay balanced", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		for i := 0; i < 20; i++ {
			mem.insertSegment(i*10, []byte{1})
		}
		assert.True(t, isBalanced(mem.root))
	})

	t.Run("tree stays balanced after removes", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		for i := 0; i < 10; i++ {
			mem.insertSegment(i*10, []byte{1})
		}
		// remove every other segment
		for i := 0; i < 10; i += 2 {
			segs := mem.segmentsForRange(i*10, i*10+1)
			if len(segs) > 0 {
				mem.removeSegment(segs[0])
			}
		}
		assert.True(t, isBalanced(mem.root))
	})
}

// read returns all bytes stored in the tree for [offset, offset+size).
// Bytes with no backing segment are returned as zero.
func read(mem *SegmentedMemory, offset, size int) []byte {
	out := make([]byte, size)
	for i := 0; i < size; i++ {
		segs := mem.segmentsForRange(offset+i, offset+i+1)
		if len(segs) == 0 {
			continue
		}
		seg := segs[len(segs)-1]
		out[i] = seg.data[(offset+i)-seg.offset]
	}
	return out
}

func TestStore(t *testing.T) {
	t.Run("all-zero data creates no segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, 4)
		mem.Store(0, []byte{0, 0, 0, 0, 0, 0, 0, 0})
		assert.Nil(t, mem.root)
	})

	t.Run("single non-overlapping store creates one segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		assert.NotNil(t, mem.root)
		assert.Equal(t, 0, mem.root.offset)
		assert.Equal(t, 4, mem.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, read(mem, 0, 4))
	})

	t.Run("store at non-zero offset records correct data", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(10, []byte{5, 6, 7, 8})
		assert.Equal(t, 10, mem.root.offset)
		assert.Equal(t, 14, mem.root.end)
		assert.Equal(t, []byte{5, 6, 7, 8}, read(mem, 10, 4))
	})

	t.Run("two non-overlapping stores create two segments", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		mem.Store(100, []byte{5, 6, 7, 8})
		assert.Equal(t, []byte{1, 2, 3, 4}, read(mem, 0, 4))
		assert.Equal(t, []byte{5, 6, 7, 8}, read(mem, 100, 4))
	})

	t.Run("store extending an existing segment merges into one", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		mem.Store(4, []byte{9, 10, 11, 12})
		assert.Nil(t, mem.root.left)
		assert.Nil(t, mem.root.right)
		assert.Equal(t, 0, mem.root.offset)
		assert.Equal(t, 8, mem.root.end)
		// original bytes before overlap are intact
		assert.Equal(t, []byte{1, 2, 3, 4}, read(mem, 0, 4))
		// overlapping region is overwritten by the second store
		assert.Equal(t, []byte{9, 10, 11, 12}, read(mem, 4, 4))
	})

	t.Run("store spanning two existing segments merges all three", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		mem.Store(8, []byte{5, 6, 7, 8})
		mem.Store(2, []byte{10, 11, 12, 13, 14, 15, 16, 17})
		segs := mem.segmentsForRange(0, 1)
		assert.Len(t, segs, 1)
		assert.Equal(t, []byte{10, 11, 12, 13, 14, 15, 16, 17}, read(mem, 2, 8))
	})

	t.Run("zero gap above threshold splits into two segments", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, 4)
		mem.Store(0, []byte{1, 2, 3, 4, 0, 0, 0, 0, 5, 6, 7, 8})
		left := mem.segmentsForRange(0, 1)
		right := mem.segmentsForRange(8, 9)
		assert.Len(t, left, 1)
		assert.Len(t, right, 1)
		assert.NotSame(t, left[0], right[0])
		assert.Equal(t, []byte{1, 2, 3, 4}, read(mem, 0, 4))
		assert.Equal(t, []byte{5, 6, 7, 8}, read(mem, 8, 4))
	})

	t.Run("zero gap below threshold does not split", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 0, 0, 0, 4, 5, 6})
		assert.NotNil(t, mem.root)
		segs := mem.segmentsForRange(0, 1)
		assert.Len(t, segs, 1)
		assert.Equal(t, 0, segs[0].offset)
		assert.Equal(t, 9, segs[0].end)
	})

	t.Run("overwriting same offset updates data in place", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		mem.Store(0, []byte{9, 8, 7, 6})
		assert.Equal(t, []byte{9, 8, 7, 6}, read(mem, 0, 4))
	})
}

func TestLoad(t *testing.T) {
	t.Run("range with no segments returns zeroes", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		result := mem.Load(0, 4)
		assert.Equal(t, []byte{0, 0, 0, 0}, result)
	})

	t.Run("load exact segment range", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.Load(0, 4))
	})

	t.Run("load subset of a segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		assert.Equal(t, []byte{3, 4, 5}, mem.Load(2, 3))
	})

	t.Run("load at non-zero offset", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(10, []byte{1, 2, 3, 4})
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.Load(10, 4))
	})

	t.Run("load spanning a segment and an empty gap returns zeroes in gap", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		result := mem.Load(0, 8)
		assert.Equal(t, []byte{1, 2, 3, 4, 0, 0, 0, 0}, result)
	})

	t.Run("load spanning two segments stitches both", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		mem.Store(4, []byte{5, 6, 7, 8})
		assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, mem.Load(0, 8))
	})

	t.Run("load spanning two segments with a gap returns zeroes in between", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, 4)
		// threshold=4: the 4 zeros split into two segments at [0,4) and [8,12)
		mem.Store(0, []byte{1, 2, 3, 4, 0, 0, 0, 0, 5, 6, 7, 8})
		result := mem.Load(0, 12)
		assert.Equal(t, []byte{1, 2, 3, 4, 0, 0, 0, 0, 5, 6, 7, 8}, result)
	})

	t.Run("load partially overlapping segment start", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(4, []byte{10, 20, 30, 40})
		// load starts before the segment
		result := mem.Load(2, 6)
		assert.Equal(t, []byte{0, 0, 10, 20, 30, 40}, result)
	})

	t.Run("load partially overlapping segment end", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{10, 20, 30, 40})
		// load ends after the segment
		result := mem.Load(2, 6)
		assert.Equal(t, []byte{30, 40, 0, 0, 0, 0}, result)
	})

	t.Run("load returns size-length slice even with no data", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		result := mem.Load(100, 8)
		assert.Len(t, result, 8)
		assert.Equal(t, make([]byte, 8), result)
	})
}

func TestClone(t *testing.T) {
	t.Run("clone of empty memory is empty", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		clone := mem.Clone().(*SegmentedMemory)
		assert.Nil(t, clone.root)
		assert.Equal(t, mem.numPages, clone.numPages)
		assert.Equal(t, mem.maxPages, clone.maxPages)
		assert.Equal(t, mem.chunkThreshold, clone.chunkThreshold)
	})

	t.Run("clone preserves stored data", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		clone := mem.Clone().(*SegmentedMemory)
		assert.Equal(t, []byte{1, 2, 3, 4}, clone.Load(0, 4))
	})

	t.Run("clone preserves multiple segments", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, 4)
		mem.Store(0, []byte{1, 2, 3, 4, 0, 0, 0, 0, 5, 6, 7, 8})
		clone := mem.Clone().(*SegmentedMemory)
		assert.Equal(t, []byte{1, 2, 3, 4}, clone.Load(0, 4))
		assert.Equal(t, []byte{5, 6, 7, 8}, clone.Load(8, 4))
	})

	t.Run("clone preserves numPages and maxPages", func(t *testing.T) {
		mem := NewSegmentedMemory(3, 10, chunkThreshold)
		clone := mem.Clone().(*SegmentedMemory)
		assert.Equal(t, 3, clone.NumPages())
		assert.Equal(t, 10, clone.MaxPages())
	})

	t.Run("writes to clone do not affect original", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		clone := mem.Clone().(*SegmentedMemory)
		clone.Store(0, []byte{9, 8, 7, 6})
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.Load(0, 4))
		assert.Equal(t, []byte{9, 8, 7, 6}, clone.Load(0, 4))
	})

	t.Run("writes to original do not affect clone", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.Store(0, []byte{1, 2, 3, 4})
		clone := mem.Clone().(*SegmentedMemory)
		mem.Store(0, []byte{9, 8, 7, 6})
		assert.Equal(t, []byte{1, 2, 3, 4}, clone.Load(0, 4))
		assert.Equal(t, []byte{9, 8, 7, 6}, mem.Load(0, 4))
	})

	t.Run("clone tree is independently balanced", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		for i := 0; i < 10; i++ {
			mem.insertSegment(i*10, []byte{1})
		}
		clone := mem.Clone().(*SegmentedMemory)
		assert.True(t, isBalanced(clone.root))
	})

	t.Run("clone parent pointers are correct", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1})
		mem.insertSegment(10, []byte{1})
		mem.insertSegment(20, []byte{1})
		clone := mem.Clone().(*SegmentedMemory)
		assert.Nil(t, clone.root.parent)
		assert.Same(t, clone.root, clone.root.left.parent)
		assert.Same(t, clone.root, clone.root.right.parent)
	})

	t.Run("clone segment data is a deep copy", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		clone := mem.Clone().(*SegmentedMemory)
		// mutate original segment's backing array directly
		mem.root.data[0] = 99
		assert.Equal(t, byte(1), clone.root.data[0])
	})
}

func TestExtendSegment(t *testing.T) {
	t.Run("no-op when range already fits inside segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		seg := mem.root
		mem.extendSegment(seg, 1, 3)
		assert.Equal(t, 0, mem.root.offset)
		assert.Equal(t, 4, mem.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.root.data)
	})

	t.Run("extends right side only", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		seg := mem.root
		mem.extendSegment(seg, 0, 8)
		assert.Equal(t, 0, seg.offset)
		assert.Equal(t, 8, seg.end)
		assert.Equal(t, 8, seg.size)
		// original data preserved at start
		assert.Equal(t, []byte{1, 2, 3, 4, 0, 0, 0, 0}, seg.data)
	})

	t.Run("extends left side only", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(4, []byte{1, 2, 3, 4})
		seg := mem.root
		mem.extendSegment(seg, 0, 8)
		assert.Equal(t, 0, seg.offset)
		assert.Equal(t, 8, seg.end)
		assert.Equal(t, 8, seg.size)
		// original data preserved at correct position in new buffer
		assert.Equal(t, []byte{0, 0, 0, 0, 1, 2, 3, 4}, seg.data)
	})

	t.Run("extends both sides", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(4, []byte{1, 2, 3, 4})
		seg := mem.root
		mem.extendSegment(seg, 2, 10)
		assert.Equal(t, 2, seg.offset)
		assert.Equal(t, 10, seg.end)
		assert.Equal(t, 8, seg.size)
		// original data sits at offset 4, which is index 2 in the new buffer
		assert.Equal(t, []byte{0, 0, 1, 2, 3, 4, 0, 0}, seg.data)
	})

	t.Run("segment remains findable in BST after left extension", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(4, []byte{1, 2, 3, 4})
		seg := mem.root
		mem.extendSegment(seg, 0, 8)
		found := mem.segmentsForRange(0, 8)
		assert.Len(t, found, 1)
		assert.Same(t, seg, found[0])
	})

	t.Run("tree stays balanced after left extension triggers re-insert", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		for i := 0; i < 7; i++ {
			mem.insertSegment(i*10+5, []byte{1})
		}
		seg := mem.segmentsForRange(5, 6)[0]
		mem.extendSegment(seg, 0, 6)
		assert.True(t, isBalanced(mem.root))
	})
}

func TestInsertSegment(t *testing.T) {
	t.Run("root segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})

		assert.NotNil(t, mem.root)
		assert.Equal(t, 0, mem.root.offset)
		assert.Equal(t, 4, mem.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.root.data)
	})

	t.Run("non-overlapping segment", func(t *testing.T) {
		mem := NewSegmentedMemory(numPages, maxPages, chunkThreshold)
		mem.insertSegment(0, []byte{1, 2, 3, 4})
		mem.insertSegment(10, []byte{5, 6, 7, 8})

		assert.NotNil(t, mem.root)
		assert.Equal(t, 0, mem.root.offset)
		assert.Equal(t, 4, mem.root.end)
		assert.Equal(t, []byte{1, 2, 3, 4}, mem.root.data)

		assert.NotNil(t, mem.root.right)
		assert.Equal(t, 10, mem.root.right.offset)
		assert.Equal(t, 14, mem.root.right.end)
		assert.Equal(t, []byte{5, 6, 7, 8}, mem.root.right.data)
	})
}
