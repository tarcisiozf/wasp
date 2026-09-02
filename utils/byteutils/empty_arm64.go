//go:build arm64

package byteutils

import (
	"unsafe"

	"github.com/jairad26/go-simd/simd_uint8"
)

const simdBlockSize = 16

func IsEmpty(bytes []byte) bool {
	size := len(bytes)
	if size == 0 {
		return true
	}
	if unsafe.SliceData(bytes) == nil {
		return true
	}
	var sum uint16
	for offset := 0; offset < size; offset += simdBlockSize {
		sum = simd_uint8.SumVecSIMD(
			&bytes[offset],
			min(simdBlockSize, size-offset),
		)
		if sum != 0 {
			return false
		}
	}
	return true
}
