package funcs

import (
	"wasp/wasp/internal/iterator"
)

type Function struct {
	Signature Signature
	Body      *iterator.Iterator
}
