package sparse_test

import (
	"testing"

	"github.com/tarcisiozf/wasp/internal/memory/sparse"
)

const pageSize = 128

func TestNewMemoryWithData(t *testing.T) {
	t.Run("empty data returns zeroed memory", func(t *testing.T) {
		mem := sparse.NewMemoryWithData(1, 10, pageSize, []byte{})
		result := mem.Load(0, pageSize)
		for i, b := range result {
			if b != 0 {
				t.Errorf("expected zero at index %d, got %d", i, b)
			}
		}
	})

	t.Run("single page of data is stored and loaded correctly", func(t *testing.T) {
		data := make([]byte, pageSize)
		for i := range data {
			data[i] = byte(i + 1)
		}
		mem := sparse.NewMemoryWithData(1, 10, pageSize, data)
		result := mem.Load(0, pageSize)
		if err := equal(data, result); err != nil {
			t.Errorf("single page round-trip failed: %v", err)
		}
	})

	t.Run("multi-page data is stored and loaded correctly", func(t *testing.T) {
		numPages := 4
		data := make([]byte, numPages*pageSize)
		for i := range data {
			data[i] = byte(i%200 + 1)
		}
		mem := sparse.NewMemoryWithData(numPages, 10, pageSize, data)
		result := mem.Load(0, numPages*pageSize)
		if err := equal(data, result); err != nil {
			t.Errorf("multi-page round-trip failed: %v", err)
		}
	})

	t.Run("sparse data with zero pages is stored and loaded correctly", func(t *testing.T) {
		// first and third page have data, second page is all zeros
		data := make([]byte, 3*pageSize)
		for i := 0; i < pageSize; i++ {
			data[i] = byte(i + 1)
		}
		// second page: all zeros (sparse)
		for i := 0; i < pageSize; i++ {
			data[2*pageSize+i] = byte(i + 50)
		}
		mem := sparse.NewMemoryWithData(3, 10, pageSize, data)
		result := mem.Load(0, 3*pageSize)
		if err := equal(data, result); err != nil {
			t.Errorf("sparse data round-trip failed: %v", err)
		}
	})

	t.Run("data written at non-zero offset loads correctly", func(t *testing.T) {
		data := make([]byte, 2*pageSize)
		for i := pageSize; i < 2*pageSize; i++ {
			data[i] = byte(i)
		}
		mem := sparse.NewMemoryWithData(2, 10, pageSize, data)
		result := mem.Load(pageSize, pageSize)
		if err := equal(data[pageSize:], result); err != nil {
			t.Errorf("offset load failed: %v", err)
		}
	})
}
