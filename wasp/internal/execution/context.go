package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/memory"
)

// BlockFrame stores runtime info about a control structure for branching
type BlockFrame struct {
	StartPos int // Position after block header (key into Blocks map)
}

type Context struct {
	Stack    *memory.Stack[any]
	Locals   *memory.Stack[any]
	Globals  *memory.Global
	Memories []*memory.Memory
	Tables   []memory.Table

	NumParams  int
	NumResults int
	Params     []any

	Body                *binary.Iterator
	FunctionCallRequest int
	Done                bool
	TailCall            bool

	Condition bool
	// BlockStack tracks nested fnblock for branching (just the start positions)
	BlockStack *memory.Stack[BlockFrame]
	// Blocks is the precomputed block target map from the function
	Blocks map[int]fnblock.Target
}

func (c *Context) Results() []any {
	return c.Stack.Last(c.NumResults)
}
