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
