package sparse_test

import (
	"testing"

	"github.com/tarcisiozf/wasp/memory/sparse"
)

const pageSize = 128

func TestDeletePageIfEmpty(t *testing.T) {
	t.Run("page with data is not deleted when overwritten with non-empty bytes", func(t *testing.T) {
		mem := sparse.NewMemory(1, 10, pageSize)

		data := make([]byte, pageSize)
		for i := range data {
			data[i] = byte(i + 1)
		}
		mem.Store(0, data)

		// overwrite with different non-empty data
		newData := make([]byte, pageSize)
		for i := range newData {
			newData[i] = byte(i + 2)
		}
		mem.Store(0, newData)

		if mem.Size() != pageSize {
			t.Errorf("expected page to still exist, Size()=%d", mem.Size())
		}
		result := mem.Load(0, pageSize)
		if err := equal(newData, result); err != nil {
			t.Errorf("data mismatch after non-empty overwrite: %v", err)
		}
	})

	t.Run("page with data is deleted when entirely overwritten with zeros", func(t *testing.T) {
		mem := sparse.NewMemory(1, 10, pageSize)

		data := make([]byte, pageSize)
		for i := range data {
			data[i] = byte(i + 1)
		}
		mem.Store(0, data)

		if mem.Size() != pageSize {
			t.Errorf("expected page to be allocated after store, Size()=%d", mem.Size())
		}

		// overwrite the full page with zeros
		zeros := make([]byte, pageSize)
		mem.Store(0, zeros)

		if mem.Size() != 0 {
			t.Errorf("expected page to be freed after zeroing, Size()=%d", mem.Size())
		}
	})

	t.Run("storing zeros to a page that never had data does not allocate a page", func(t *testing.T) {
		mem := sparse.NewMemory(1, 10, pageSize)

		zeros := make([]byte, pageSize)
		mem.Store(0, zeros)

		if mem.Size() != 0 {
			t.Errorf("expected no page to be allocated for all-zero store, Size()=%d", mem.Size())
		}
	})

	t.Run("partial zero overwrite of a page does not delete the page if data remains", func(t *testing.T) {
		mem := sparse.NewMemory(1, 10, pageSize)

		data := make([]byte, pageSize)
		for i := range data {
			data[i] = byte(i + 1)
		}
		mem.Store(0, data)

		// zero out only the first half
		zeros := make([]byte, pageSize/2)
		mem.Store(0, zeros)

		if mem.Size() != pageSize {
			t.Errorf("expected page to remain because second half still has data, Size()=%d", mem.Size())
		}

		result := mem.Load(0, pageSize)
		for i := 0; i < pageSize/2; i++ {
			if result[i] != 0 {
				t.Errorf("expected zero at index %d, got %d", i, result[i])
			}
		}
		for i := pageSize / 2; i < pageSize; i++ {
			if result[i] != data[i] {
				t.Errorf("expected original data at index %d, got %d (want %d)", i, result[i], data[i])
			}
		}
	})

	t.Run("zeroing one page does not affect adjacent pages", func(t *testing.T) {
		mem := sparse.NewMemory(2, 10, pageSize)

		page0 := make([]byte, pageSize)
		page1 := make([]byte, pageSize)
		for i := range page0 {
			page0[i] = byte(i + 1)
			page1[i] = byte(i + 100)
		}
		mem.Store(0, page0)
		mem.Store(pageSize, page1)

		if mem.Size() != 2*pageSize {
			t.Errorf("expected 2 pages allocated, Size()=%d", mem.Size())
		}

		// zero out page 0
		mem.Store(0, make([]byte, pageSize))

		if mem.Size() != pageSize {
			t.Errorf("expected 1 page remaining after zeroing page 0, Size()=%d", mem.Size())
		}

		result := mem.Load(pageSize, pageSize)
		if err := equal(page1, result); err != nil {
			t.Errorf("page 1 data corrupted after zeroing page 0: %v", err)
		}
	})
}

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
