package baz_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/tarcisiozf/wasp/internal/baz"
	"github.com/tarcisiozf/wasp/internal/memory"
)

var rng = rand.New(rand.NewSource(3847234571923791))

const (
	numCases = 1000
)

type storeTestCase struct {
	offset int
	data   []byte
}

type loadTestCase struct {
	offset int
	size   int
}

var storeTestCases = make([]storeTestCase, numCases)
var loadTestCases = make([]loadTestCase, numCases)

func init() {
	for i := 0; i < numCases; i++ {
		storeTestCases[i] = storeTestCase{
			offset: rng.Intn(65536),
			data:   randomBytes(rng.Intn(1024)),
		}
	}
	for i := 0; i < numCases; i++ {
		loadTestCases[i] = loadTestCase{
			offset: rng.Intn(65536),
			size:   rng.Intn(2048),
		}
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
	mem := memory.NewMemory(10, 0)
	bmem := baz.New(0, 0)

	for i := 0; i < numCases; i++ {
		storeCase := storeTestCases[i]
		mem.Store(storeCase.offset, storeCase.data)
		bmem.Store(storeCase.offset, storeCase.data)

		loadCase := loadTestCases[i]
		expected := mem.Load(loadCase.offset, loadCase.size)
		actual := bmem.Load(loadCase.offset, loadCase.size)

		if err := equal(expected, actual); err != nil {
			t.Fatalf("Test case %d failed: %v", i, err)
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
