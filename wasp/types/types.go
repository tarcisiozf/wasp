package types

type Type struct {
	Code byte
	x    any
}

func (t *Type) Zero() any {
	switch t.x.(type) {
	case int32:
		return int32(0)
	default:
		panic("unsupported type")
	}
}

var types = make([]Type, 256)

func newType[T any](code byte) Type {
	var zero T
	t := Type{Code: code, x: zero}
	types[code] = t
	return t
}

var (
	Int32 = newType[int32](0x7F)
)

func ForCode(code byte) Type {
	return types[code]
}
