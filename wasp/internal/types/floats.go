package types

import (
	"reflect"
	"wasp/wasp/internal/binary"
)

type typeFloat32 struct{}

func (t typeFloat32) String() string {
	return "float32"
}

func (t typeFloat32) Zero() any {
	return float32(0)
}

func (t typeFloat32) Read(iterator *binary.Iterator) any {
	return iterator.Float32()
}

func (t typeFloat32) Kind() reflect.Kind {
	return reflect.Float32
}

type typeFloat64 struct{}

func (t typeFloat64) String() string {
	return "float64"
}

func (t typeFloat64) Zero() any {
	return float64(0)
}

func (t typeFloat64) Read(iterator *binary.Iterator) any {
	return iterator.Float64()
}

func (t typeFloat64) Kind() reflect.Kind {
	return reflect.Float64
}
