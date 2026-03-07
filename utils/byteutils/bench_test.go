package byteutils_test

import (
	"testing"

	"github.com/tarcisiozf/wasp/utils/byteutils"
)

func BenchmarkIsEmpty(b *testing.B) {
	// Create a large byte slice filled with zeros
	size := 1024 * 1024 // 1 MB
	emptyBytes := make([]byte, size)

	// Create a large byte slice filled with non-zero values
	nonEmptyBytes := make([]byte, size)
	for i := range nonEmptyBytes {
		nonEmptyBytes[i] = byte(i%256 + 1) // Fill with non-zero values
	}

	b.Run("Empty Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if !byteutils.IsEmpty(emptyBytes) {
				b.Fatal("Expected empty slice to be detected as empty")
			}
		}
	})

	b.Run("Non-Empty Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if byteutils.IsEmpty(nonEmptyBytes) {
				b.Fatalf("[%d] Expected non-empty slice to be detected as non-empty", i)
			}
		}
	})
}
