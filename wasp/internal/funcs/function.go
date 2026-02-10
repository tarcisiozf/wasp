package funcs

import (
	"wasp/wasp/internal/binary"
)

type Function struct {
	Signature Signature
	Body      *binary.Iterator
}
