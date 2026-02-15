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

	NumParams  int
	NumResults int
	Params     []any
	Results    []any

	Body                *binary.Iterator
	FunctionCallRequest int
	Done                bool

	Condition bool
	// BlockStack tracks nested fnblock for branching (just the start positions)
	BlockStack *memory.Stack[BlockFrame]
	// Blocks is the precomputed block target map from the function
	Blocks map[int]fnblock.Target
}
