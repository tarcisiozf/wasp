package types

import (
	"fmt"
	"reflect"

	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/opcodes"
)

type asType interface {
	Zero() any
	Read(*binary.Iterator) any
	Kind() reflect.Kind
	String() string
}

type Type struct {
	asType

	Code  byte
	Const uint16
}

var types = make([]Type, 256)

func addType(code byte, typeConst uint16, t asType) Type {
	types[code] = Type{asType: t, Code: code, Const: typeConst}
	return types[code]
}

var (
	Int32   = addType(0x7F, opcodes.I32Const, &typeInt32{})
	Int64   = addType(0x7E, opcodes.I64Const, &typeInt64{})
	Float32 = addType(0x7D, opcodes.F32Const, &typeFloat32{})
	Float64 = addType(0x7C, opcodes.F64Const, &typeFloat64{})
	Void    = addType(0x40, 0, &typeVoid{})
)

func ForCode(code byte) Type {
	t := types[code]
	if t.Code == 0 {
		panic(fmt.Sprintf("invalid type code: 0x%x", code))
	}
	return t
}
