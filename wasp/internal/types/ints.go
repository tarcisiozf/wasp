package types

import (
	"reflect"
	"wasp/wasp/internal/binary"
)

type typeInt32 struct{}

func (t typeInt32) String() string {
	return "int32"
}

func (t typeInt32) Zero() any {
	return int32(0)
}

func (t typeInt32) Read(iter *binary.Iterator) any {
	return int32(iter.Varint())
}

func (t typeInt32) Kind() reflect.Kind {
	return reflect.Int32
}

type typeInt64 struct{}

func (t typeInt64) String() string {
	return "int64"
}

func (t typeInt64) Zero() any {
	return int64(0)
}

func (t typeInt64) Read(iter *binary.Iterator) any {
	return int64(iter.Varint())
}

func (t typeInt64) Kind() reflect.Kind {
	return reflect.Int64
}
