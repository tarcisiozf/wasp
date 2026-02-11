package types

import (
	"reflect"
	"wasp/wasp/internal/binary"
)

type Type interface {
	Code() byte
	Zero() any
	Read(*binary.Iterator) any
	Kind() reflect.Kind
}

var types = make([]Type, 256)

func addType(t Type) Type {
	types[t.Code()] = t
	return t
}

var (
	Int32 = addType(&typeInt32{})
)

func ForCode(code byte) Type {
	return types[code]
}
