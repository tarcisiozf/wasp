package memory

import "errors"

var (
	ErrIndexOutOfBounds         = errors.New("index out of bounds")
	ErrCannotSetImmutableGlobal = errors.New("cannot set immutable global")
)
