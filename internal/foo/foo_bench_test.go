package foo_test

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"testing"

	"github.com/tarcisiozf/wasp/internal/bar"
	"github.com/tarcisiozf/wasp/internal/foo"
)

const (
	numCases = 1000
)

var (
	avlmem = bar.NewSegmentedMemory(65536, 65536, 256)
	slmem  = foo.New(0, 0)
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
			offset: rand.Intn(65536),
			data:   randomBytes(rand.Intn(1024)),
		}
		//fmt.Printf("Generated store test case %d: offset=%d, size=%d\n", i, storeTestCases[i].offset, len(storeTestCases[i].data))
	}
	for i := 0; i < numCases; i++ {
		loadTestCases[i] = LoadTestCase{
			offset: rand.Intn(65536),
			size:   rand.Intn(2048),
		}
		//fmt.Printf("Generated load test case %d: offset=%d, size=%d\n", i, loadTestCases[i].offset, loadTestCases[i].size)
	}
	fmt.Printf("Generated %d store and load test cases\n", numCases)
}

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := crand.Read(bytes)
	if err != nil {
		panic("failed to generate random bytes: " + err.Error())
	}
	return bytes
}

func BenchmarkAvlStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := storeTestCases[i%numCases]
		avlmem.Store(tc.offset, tc.data)
	}
}

func BenchmarkSkipStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := storeTestCases[i%numCases]
		slmem.Store(tc.offset, tc.data)
	}
}

func BenchmarkAvlLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := loadTestCases[i%numCases]
		_ = avlmem.Load(tc.offset, tc.size)
	}
}

func BenchmarkSkipLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := loadTestCases[i%numCases]
		_ = slmem.Load(tc.offset, tc.size)
	}
}
