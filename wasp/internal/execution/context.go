package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Local               []any
	Stack               *memory.Stack[any]
	Body                *binary.Iterator
	Done                bool
	FunctionCallRequest int
	Globals             *memory.Global
}
