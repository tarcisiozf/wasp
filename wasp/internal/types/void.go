package types

import (
	"reflect"
	"wasp/wasp/internal/binary"
)

type typeVoid struct{}

func (t typeVoid) Zero() any {
	panic("cannot get zero value of void type")
}

func (t typeVoid) Read(*binary.Iterator) any {
	panic("cannot read void type")
}

func (t typeVoid) Kind() reflect.Kind {
	return reflect.Invalid
}
