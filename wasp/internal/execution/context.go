package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/memory"
)

type Context struct {
	Stack    *memory.Stack[any]
	Locals   *memory.Stack[any]
	Globals  *memory.Global
	Memories []*memory.Memory

	Body *binary.Iterator

	FunctionCallRequest int
	BlockType           byte
	Done                bool
	Depth               int
	Condition           bool
}
