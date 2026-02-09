package execution

import (
	"wasp/wasp/internal/iterator"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Local []any
	Stack *memory.Stack
	Body  *iterator.Iterator
	Done  bool
}
