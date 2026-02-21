package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/funcs/fnblock"
	"wasp/wasp/internal/funcs/fnsig"
	"wasp/wasp/internal/memory"
)

// BlockFrame stores runtime info about a control structure for branching
type BlockFrame struct {
	StartPos int // Position after block header (key into Blocks map)
}

type Context struct {
	Globals        *memory.Global
	Memories       []*memory.Memory
	Tables         []*memory.Table
	FuncSignatures []fnsig.Signature // indexed by function index
	TypeSignatures []fnsig.Signature // indexed by type index (for call_indirect)

	Stack  *memory.Stack[any]
	Locals *memory.Stack[any]

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

	Debug bool
}

func (c *Context) Results() []any {
	return c.Stack.Last(c.NumResults)
}
