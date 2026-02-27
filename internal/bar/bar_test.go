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
