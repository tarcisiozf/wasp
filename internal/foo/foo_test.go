package foo

import (
	"testing"
)

func newMem(numPages int) *FragmentedMemory {
	return NewFragmentedMemory(numPages, 0)
}

func TestStore_Load_RoundTrip(t *testing.T) {
	mem := newMem(1)
	data := []byte{1, 2, 3, 4, 5}
	mem.Store(0, data)
	got := mem.Load(0, len(data))
	for i, b := range data {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestStore_Load_Offset(t *testing.T) {
	mem := newMem(1)
	data := []byte{10, 20, 30}
	mem.Store(100, data)
	got := mem.Load(100, len(data))
	for i, b := range data {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestLoad_UninitializedReturnsZeros(t *testing.T) {
	mem := newMem(1)
	got := mem.Load(0, 4)
	for i, b := range got {
		if b != 0 {
			t.Errorf("byte %d: want 0, got %d", i, b)
		}
	}
}

func TestStore_OverwriteExistingSegment(t *testing.T) {
	mem := newMem(1)
	mem.Store(0, []byte{1, 2, 3, 4, 5})
	mem.Store(1, []byte{99, 98})
	got := mem.Load(0, 5)
	want := []byte{1, 99, 98, 4, 5}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestStore_NonContiguousSegments(t *testing.T) {
	mem := newMem(1)
	mem.Store(0, []byte{1, 2})
	mem.Store(10, []byte{3, 4})

	first := mem.Load(0, 2)
	if first[0] != 1 || first[1] != 2 {
		t.Errorf("first segment: want [1 2], got %v", first)
	}

	second := mem.Load(10, 2)
	if second[0] != 3 || second[1] != 4 {
		t.Errorf("second segment: want [3 4], got %v", second)
	}
}

func TestStore_OutOfBounds_Panics(t *testing.T) {
	mem := newMem(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-bounds store")
		}
	}()
	mem.Store(mem.numPages*pageSize, []byte{1})
}

func TestLoad_OutOfBounds_Panics(t *testing.T) {
	mem := newMem(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-bounds load")
		}
	}()
	mem.Load(mem.numPages*pageSize, 1)
}

func TestStore_NegativeOffset_Panics(t *testing.T) {
	mem := newMem(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative offset store")
		}
	}()
	mem.Store(-1, []byte{1})
}

func TestLoad_NegativeOffset_Panics(t *testing.T) {
	mem := newMem(1)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative offset load")
		}
	}()
	mem.Load(-1, 1)
}

// ---- helpers ----

func sequentialStores(mem interface {
	Store(int, []byte)
}, n, chunkSize int) {
	chunk := make([]byte, chunkSize)
	for i := range chunk {
		chunk[i] = byte(i)
	}
	for i := 0; i < n; i++ {
		mem.Store(i*chunkSize, chunk)
	}
}

// ---- benchmarks ----

const (
	benchChunkSize = 4
	benchChunks    = 256 // 256 * 4 = 1 KiB total
)

func BenchmarkStore(b *testing.B) {
	for b.Loop() {
		mem := NewFragmentedMemory(1, 0)
		sequentialStores(mem, benchChunks, benchChunkSize)
	}
}

func BenchmarkLoad(b *testing.B) {
	mem := NewFragmentedMemory(1, 0)
	sequentialStores(mem, benchChunks, benchChunkSize)
	total := benchChunks * benchChunkSize
	b.ResetTimer()
	for b.Loop() {
		_ = mem.Load(0, total)
	}
}

func BenchmarkStoreAndLoad(b *testing.B) {
	total := benchChunks * benchChunkSize
	for b.Loop() {
		mem := NewFragmentedMemory(1, 0)
		sequentialStores(mem, benchChunks, benchChunkSize)
		_ = mem.Load(0, total)
	}
}
