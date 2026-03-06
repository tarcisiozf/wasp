package intervaltree

import (
	"fmt"
	"math/rand"
	"testing"
)

var rng = rand.New(rand.NewSource(2472782345746534))

const (
	numCases = 1000
)

var (
	mem = NewMemory(65536, 65536, 256)
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

func BenchmarkStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := storeTestCases[i%numCases]
		mem.Store(tc.offset, tc.data)
	}
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := loadTestCases[i%numCases]
		_ = mem.Load(tc.offset, tc.size)
	}
}
