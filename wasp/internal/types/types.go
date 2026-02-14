package types

import (
	"fmt"
	"reflect"
	"wasp/wasp/internal/binary"
)

type asType interface {
	Zero() any
	Read(*binary.Iterator) any
	Kind() reflect.Kind
}

type Type struct {
	asType

	Code byte
}

var types = make([]Type, 256)

func addType(code byte, t asType) Type {
	types[code] = Type{asType: t, Code: code}
	return types[code]
}

var (
	Int32   = addType(0x7F, &typeInt32{})
	Int64   = addType(0x7E, &typeInt64{})
	Float32 = addType(0x7D, &typeFloat32{})
	Float64 = addType(0x7C, &typeFloat64{})
	Void    = addType(0x40, &typeVoid{})
)

func ForCode(code byte) Type {
	t := types[code]
	if t.Code == 0 {
		panic(fmt.Sprintf("invalid type code: 0x%x", code))
	}
	return t
}
