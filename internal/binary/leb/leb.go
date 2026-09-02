package leb

import (
	"fmt"
	"io"
)

func Encode(value any) []byte {
	switch value.(type) {
	case int:
		return EncodeInt(value.(int))
	}
	panic(fmt.Sprintf("unsupported type %T", value))
}

func EncodeInt(v int) []byte {
	var bytes []byte

	for {
		b := byte(v & 0x7F)
		v >>= 7
		if (v == 0 && (b&0x40) == 0) || (v == -1 && (b&0x40) != 0) {
			bytes = append(bytes, b)
			break
		}
		bytes = append(bytes, b|0x80)
	}

	return bytes
}

func WriteUint(w io.ByteWriter, v uint64) int {
	i := 0
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v == 0 {
			if err := w.WriteByte(b); err != nil {
				panic(fmt.Sprintf("failed to write byte: %v", err))
			}
			i++
			break
		}
		if err := w.WriteByte(b | 0x80); err != nil {
			panic(fmt.Sprintf("failed to write byte: %v", err))
		}
		i++
	}
	return i
}
