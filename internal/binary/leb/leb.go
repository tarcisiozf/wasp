package leb

import "fmt"

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

func EncodeUint(buf []byte, v uint64) int {
	i := 0
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v == 0 {
			buf[i] = b
			i++
			break
		}
		buf[i] = b | 0x80
		i++
	}
	return i
}
