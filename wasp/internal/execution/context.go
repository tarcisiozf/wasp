package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Local               []any
	Stack               *memory.Stack
	Body                *binary.Iterator
	Done                bool
	FunctionCallRequest int
	Globals             *memory.Global
}
