package types

import (
	"github.com/tarcisiozf/wasp/internal/binary"
	"reflect"
)

type typeVoid struct{}

func (t typeVoid) String() string {
	return "void"
}

func (t typeVoid) Zero() any {
	panic("cannot get zero value of void type")
}

func (t typeVoid) Read(*binary.Iterator) any {
	panic("cannot read void type")
}

func (t typeVoid) Kind() reflect.Kind {
	return reflect.Invalid
}
