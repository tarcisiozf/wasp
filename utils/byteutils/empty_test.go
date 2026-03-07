package byteutils_test

import (
	"testing"

	"github.com/tarcisiozf/wasp/utils/byteutils"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: true,
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: true,
		},
		{
			name:     "single zero byte",
			input:    []byte{0},
			expected: true,
		},
		{
			name:     "single non-zero byte",
			input:    []byte{1},
			expected: false,
		},
		{
			name:     "all zeros",
			input:    []byte{0, 0, 0, 0, 0},
			expected: true,
		},
		{
			name:     "first byte non-zero",
			input:    []byte{1, 0, 0, 0, 0},
			expected: false,
		},
		{
			name:     "last byte non-zero",
			input:    []byte{0, 0, 0, 0, 1},
			expected: false,
		},
		{
			name:     "middle byte non-zero",
			input:    []byte{0, 0, 1, 0, 0},
			expected: false,
		},
		{
			name:     "all non-zero",
			input:    []byte{1, 2, 3, 4, 5},
			expected: false,
		},
		{
			name:     "large empty slice",
			input:    make([]byte, 4096),
			expected: true,
		},
		{
			name:     "max byte value",
			input:    []byte{0xFF},
			expected: false,
		},
		{
			name:     "zeros with max byte at end",
			input:    []byte{0, 0, 0, 0xFF},
			expected: false,
		},
	}

	// test with large non-empty slice (non-zero at very end)
	largeNonEmpty := make([]byte, 4096)
	largeNonEmpty[4095] = 1
	tests = append(tests, struct {
		name     string
		input    []byte
		expected bool
	}{
		name:     "large slice with last byte non-zero",
		input:    largeNonEmpty,
		expected: false,
	})

	// test with large non-empty slice (non-zero at very beginning)
	largeNonEmptyStart := make([]byte, 4096)
	largeNonEmptyStart[0] = 1
	tests = append(tests, struct {
		name     string
		input    []byte
		expected bool
	}{
		name:     "large slice with first byte non-zero",
		input:    largeNonEmptyStart,
		expected: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := byteutils.IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsEmpty_VariousSizes(t *testing.T) {
	// Test with various sizes to exercise potential SIMD boundary conditions
	sizes := []int{1, 2, 3, 4, 7, 8, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 512, 1024}

	for _, size := range sizes {
		// all zeros
		zeros := make([]byte, size)
		if !byteutils.IsEmpty(zeros) {
			t.Errorf("IsEmpty(zeros of size %d) = false, want true", size)
		}

		// last byte non-zero
		nonZero := make([]byte, size)
		nonZero[size-1] = 42
		if byteutils.IsEmpty(nonZero) {
			t.Errorf("IsEmpty(non-zero at end, size %d) = true, want false", size)
		}
	}
}
