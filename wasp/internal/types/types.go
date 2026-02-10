package types

import "wasp/wasp/internal/binary"

type Type struct {
	Code byte
	Zero func() any
	Read func(*binary.Iterator) any
}

var types = make([]*Type, 256)

func newType[T any](code byte) *Type {
	var x T
	var y any = x

	var zero func() any
	var read func(*binary.Iterator) any

	switch y.(type) {
	case int32:
		zero = func() any { return int32(0) }
		read = func(it *binary.Iterator) any { return int32(it.Varint()) }
	default:
		panic("unsupported type")
	}

	t := &Type{
		Code: code,
		Zero: zero,
		Read: read,
	}
	types[code] = t
	return t
}

var (
	Int32 = newType[int32](0x7F)
)

func ForCode(code byte) *Type {
	return types[code]
}
