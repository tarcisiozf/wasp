package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Local   []any
	Stack   *memory.Stack[any]
	Globals *memory.Global

	Body *binary.Iterator

	FunctionCallRequest int
	BlockType           byte
	Done                bool
	Depth               int
	Condition           bool
}
