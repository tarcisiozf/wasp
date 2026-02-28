package baz

import (
	"fmt"
	"math/rand"
	"testing"
)

var rng = rand.New(rand.NewSource(2472782345746534))

const (
	numBenchCases = 1000
)

var (
	benchMem = New(65536, 65536)
)

type benchStoreCase struct {
	offset int
	data   []byte
}

type benchLoadCase struct {
	offset int
	size   int
}

var benchStoreCases = make([]benchStoreCase, numBenchCases)
var benchLoadCases = make([]benchLoadCase, numBenchCases)

func init() {
	for i := 0; i < numBenchCases; i++ {
		benchStoreCases[i] = benchStoreCase{
			offset: rng.Intn(65536),
			data:   benchRandomBytes(rng.Intn(1024)),
		}
	}
	for i := 0; i < numBenchCases; i++ {
		benchLoadCases[i] = benchLoadCase{
			offset: rng.Intn(65536),
			size:   rng.Intn(2048),
		}
	}
	fmt.Printf("Generated %d bench store and load test cases\n", numBenchCases)
}

func benchRandomBytes(n int) []byte {
	bytes := make([]byte, n)
	for i := range n {
		bytes[i] = byte(rng.Intn(256))
	}
	return bytes
}

func BenchmarkStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := benchStoreCases[i%numBenchCases]
		benchMem.Store(tc.offset, tc.data)
	}
}

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tc := benchLoadCases[i%numBenchCases]
		_ = benchMem.Load(tc.offset, tc.size)
	}
}
