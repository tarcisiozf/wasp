//go:build arm64

package byteutils

import "github.com/jairad26/go-simd/simd_uint8"

func IsEmpty(bytes []byte) bool {
	size := len(bytes)
	if size == 0 {
		return true
	}
	return simd_uint8.SumVecSIMD(&bytes[0], size) == 0
}
