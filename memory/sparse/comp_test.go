package sparse_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/tarcisiozf/wasp/internal/serialization"
	"github.com/tarcisiozf/wasp/memory/contiguous"
	"github.com/tarcisiozf/wasp/memory/sparse"
)

var rng = rand.New(rand.NewSource(3847234571923791))

const (
	numCases = 1000
)

type storeTestCase struct {
	offset int
	data   []byte
}

type LoadTestCase struct {
	offset int
	size   int
}

var storeTestCases = make([]storeTestCase, numCases)
var loadTestCases = make([]LoadTestCase, numCases)

func init() {
	for i := 0; i < numCases; i++ {
		storeTestCases[i] = storeTestCase{
			offset: rng.Intn(65536),
			data:   randomBytes(rng.Intn(1024)),
		}
		//fmt.Printf("Generated store test case %d: offset=%d, size=%d\n", i, storeTestCases[i].offset, len(storeTestCases[i].data))
	}
	for i := 0; i < numCases; i++ {
		loadTestCases[i] = LoadTestCase{
			offset: rng.Intn(65536),
			size:   rng.Intn(2048),
		}
		//fmt.Printf("Generated load test case %d: offset=%d, size=%d\n", i, loadTestCases[i].offset, loadTestCases[i].size)
	}
	fmt.Printf("Generated %d store and load test cases\n", numCases)
}

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	for i := range n {
		bytes[i] = byte(rng.Intn(256))
	}
	return bytes
}

func TestCompare(t *testing.T) {
	mem := contiguous.NewMemory(10, 0)
	smem := sparse.NewMemory(10, 0, 128)

	for i := 0; i < numCases; i++ {
		storeCase := storeTestCases[i]
		mem.Store(storeCase.offset, storeCase.data)
		smem.Store(storeCase.offset, storeCase.data)

		loadCase := loadTestCases[i]
		expected := mem.Load(loadCase.offset, loadCase.size)
		actual := smem.Load(loadCase.offset, loadCase.size)

		if err := equal(expected, actual); err != nil {
			t.Errorf("test case %d failed: %v", i, err)
		}
	}
}

func TestSerialization(t *testing.T) {
	smem := sparse.NewMemory(10, 0, 128)

	for i := 0; i < numCases; i++ {
		storeCase := storeTestCases[i]
		smem.Store(storeCase.offset, storeCase.data)
	}

	var buf bytes.Buffer
	enc := serialization.NewEncoder(&buf)
	sparse.Encode(enc, smem)

	encoded := buf.Bytes()
	dec := serialization.NewDecoder(bytes.NewReader(encoded))
	decodedMem, err := sparse.Decode(dec)
	if err != nil {
		t.Fatalf("decoding failed: %v", err)
	}

	for i := 0; i < numCases; i++ {
		loadCase := loadTestCases[i]
		expected := smem.Load(loadCase.offset, loadCase.size)
		actual := decodedMem.Load(loadCase.offset, loadCase.size)

		if err := equal(expected, actual); err != nil {
			t.Errorf("serialization test case %d failed: %v", i, err)
		}
	}
}

func equal(a, b []byte) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch: expected %d, got %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			return fmt.Errorf("byte mismatch at index %d: expected %d, got %d", i, a[i], b[i])
		}
	}
	return nil
}
