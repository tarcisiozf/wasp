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

// ---- clone / copy-on-write tests ----

func TestClone_ReadsMatchOriginal(t *testing.T) {
	orig := newMem(1)
	orig.Store(0, []byte{1, 2, 3, 4})
	clone := orig.Clone().(*FragmentedMemory)

	got := clone.Load(0, 4)
	want := []byte{1, 2, 3, 4}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("clone byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestClone_WriteToCloneDoesNotAffectOriginal(t *testing.T) {
	orig := newMem(1)
	orig.Store(0, []byte{1, 2, 3, 4})
	clone := orig.Clone().(*FragmentedMemory)

	clone.Store(0, []byte{99, 99, 99, 99})

	got := orig.Load(0, 4)
	want := []byte{1, 2, 3, 4}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("original byte %d: want %d, got %d (clone write leaked)", i, b, got[i])
		}
	}
}

func TestClone_WriteToOriginalDoesNotAffectClone(t *testing.T) {
	orig := newMem(1)
	orig.Store(0, []byte{1, 2, 3, 4})
	clone := orig.Clone().(*FragmentedMemory)

	orig.Store(0, []byte{99, 99, 99, 99})

	got := clone.Load(0, 4)
	want := []byte{1, 2, 3, 4}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("clone byte %d: want %d, got %d (original write leaked)", i, b, got[i])
		}
	}
}

func TestClone_SharedDataBeforeWrite(t *testing.T) {
	orig := newMem(1)
	orig.Store(0, []byte{1, 2, 3, 4})
	clone := orig.Clone().(*FragmentedMemory)

	// before any write, both trees should share the same underlying slice
	if orig.root == nil || clone.root == nil {
		t.Fatal("expected both to have a root segment")
	}
	origPtr := &orig.root.data[0]
	clonePtr := &clone.root.data[0]
	if origPtr != clonePtr {
		t.Error("expected shared backing array before any COW write")
	}

	// after a write to the clone, slices must diverge
	clone.Store(0, []byte{99})
	clonePtr = &clone.root.data[0]
	if origPtr == clonePtr {
		t.Error("expected separate backing array after COW write")
	}
}

// ---- sparse correctness tests ----

func TestSparse_SkipsLeadingAndTrailingZeros(t *testing.T) {
	mem := NewSparseMemory(1, 0, 4)
	mem.Store(0, []byte{0, 0, 1, 2, 3, 0, 0})
	got := mem.Load(0, 7)
	want := []byte{0, 0, 1, 2, 3, 0, 0}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestSparse_BridgesSmallZeroGap(t *testing.T) {
	// gap of 3 zeros ≤ threshold of 4 → single segment
	mem := NewSparseMemory(1, 0, 4)
	mem.Store(0, []byte{1, 2, 0, 0, 0, 3, 4})
	got := mem.Load(0, 7)
	want := []byte{1, 2, 0, 0, 0, 3, 4}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestSparse_SplitsLargeZeroGap(t *testing.T) {
	// gap of 5 zeros > threshold of 4 → two separate segments; zeros read back as 0
	mem := NewSparseMemory(1, 0, 4)
	mem.Store(0, []byte{1, 2, 0, 0, 0, 0, 0, 3, 4})
	got := mem.Load(0, 9)
	want := []byte{1, 2, 0, 0, 0, 0, 0, 3, 4}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

func TestSparse_AllZeros_StoresNothing(t *testing.T) {
	mem := NewSparseMemory(1, 0, 4)
	mem.Store(0, make([]byte, 64))
	if mem.root != nil {
		t.Error("expected no segments for all-zero payload")
	}
}

func TestSparse_NonSparseUnchanged(t *testing.T) {
	mem := NewFragmentedMemory(1, 0)
	mem.Store(0, []byte{1, 0, 0, 0, 0, 0, 2})
	got := mem.Load(0, 7)
	want := []byte{1, 0, 0, 0, 0, 0, 2}
	for i, b := range want {
		if got[i] != b {
			t.Errorf("byte %d: want %d, got %d", i, b, got[i])
		}
	}
}

// ---- memory usage ----

// segmentBytes returns the total number of bytes held across all segments in the tree.
func segmentBytes(root *Segment) int {
	if root == nil {
		return 0
	}
	return len(root.data) + segmentBytes(root.left) + segmentBytes(root.right)
}

// segmentCount returns the number of segments in the tree.
func segmentCount(root *Segment) int {
	if root == nil {
		return 0
	}
	return 1 + segmentCount(root.left) + segmentCount(root.right)
}

func TestSparse_MemoryUsage(t *testing.T) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	totalPayload := len(payload)

	without := NewFragmentedMemory(1, 0)
	without.Store(0, payload)

	with := NewSparseMemory(1, 0, sparseThreshold)
	with.Store(0, payload)

	bytesWithout := segmentBytes(without.root)
	bytesWith := segmentBytes(with.root)
	segsWithout := segmentCount(without.root)
	segsWith := segmentCount(with.root)
	saved := bytesWithout - bytesWith
	pct := float64(saved) / float64(bytesWithout) * 100

	t.Logf("payload size:       %d bytes", totalPayload)
	t.Logf("without flag:       %d bytes in %d segment(s)", bytesWithout, segsWithout)
	t.Logf("with flag:          %d bytes in %d segment(s)", bytesWith, segsWith)
	t.Logf("saved:              %d bytes (%.1f%%)", saved, pct)

	nonZeroBytes := sparseNonZeroChunks * sparseNonZeroSize
	if bytesWith > nonZeroBytes {
		t.Errorf("sparse stored more than the non-zero content (%d > %d)", bytesWith, nonZeroBytes)
	}
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

// sparsePayload builds a realistic sparse payload:
// numChunks non-zero regions of chunkSize bytes, each separated by gapSize zero bytes.
func sparsePayload(numChunks, chunkSize, gapSize int) []byte {
	total := numChunks*chunkSize + (numChunks-1)*gapSize
	buf := make([]byte, total)
	pos := 0
	for i := 0; i < numChunks; i++ {
		for j := 0; j < chunkSize; j++ {
			buf[pos] = byte(i*chunkSize + j + 1)
			pos++
		}
		pos += gapSize // leave zeros
	}
	return buf
}

// ---- benchmarks ----

const (
	benchChunkSize = 4
	benchChunks    = 256 // 256 * 4 = 1 KiB total

	// sparse benchmark: 64 non-zero chunks of 8 bytes, separated by 64 zero bytes
	sparseNonZeroChunks = 64
	sparseNonZeroSize   = 8
	sparseGapSize       = 64
	sparseThreshold     = 4
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

func BenchmarkSparse_Store_WithoutFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	b.ResetTimer()
	for b.Loop() {
		mem := NewFragmentedMemory(1, 0)
		mem.Store(0, payload)
	}
}

func BenchmarkSparse_Store_WithFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	b.ResetTimer()
	for b.Loop() {
		mem := NewSparseMemory(1, 0, sparseThreshold)
		mem.Store(0, payload)
	}
}

func BenchmarkSparse_Load_WithoutFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	mem := NewFragmentedMemory(1, 0)
	mem.Store(0, payload)
	b.ResetTimer()
	for b.Loop() {
		_ = mem.Load(0, len(payload))
	}
}

func BenchmarkSparse_Load_WithFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	mem := NewSparseMemory(1, 0, sparseThreshold)
	mem.Store(0, payload)
	b.ResetTimer()
	for b.Loop() {
		_ = mem.Load(0, len(payload))
	}
}

func BenchmarkSparse_StoreAndLoad_WithoutFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	total := len(payload)
	b.ResetTimer()
	for b.Loop() {
		mem := NewFragmentedMemory(1, 0)
		mem.Store(0, payload)
		_ = mem.Load(0, total)
	}
}

func BenchmarkSparse_StoreAndLoad_WithFlag(b *testing.B) {
	payload := sparsePayload(sparseNonZeroChunks, sparseNonZeroSize, sparseGapSize)
	total := len(payload)
	b.ResetTimer()
	for b.Loop() {
		mem := NewSparseMemory(1, 0, sparseThreshold)
		mem.Store(0, payload)
		_ = mem.Load(0, total)
	}
}
