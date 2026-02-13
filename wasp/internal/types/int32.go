package types

import (
	"reflect"
	"wasp/wasp/internal/binary"
)

type typeInt32 struct{}

func (t *typeInt32) Zero() any {
	return int32(0)
}

func (t *typeInt32) Read(iter *binary.Iterator) any {
	return int32(iter.Varint())
}

func (t *typeInt32) Kind() reflect.Kind {
	return reflect.Int32
}
