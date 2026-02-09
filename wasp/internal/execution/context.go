package execution

import (
	"wasp/wasp/internal/iterator"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Local []memory.Local
	Stack *memory.Stack
	Body  *iterator.Iterator
	Done  bool
}
