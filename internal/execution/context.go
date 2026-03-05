package execution

import (
	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/funcs/fnblock"
	"github.com/tarcisiozf/wasp/internal/funcs/fnsig"
	"github.com/tarcisiozf/wasp/internal/memory"
	iface "github.com/tarcisiozf/wasp/memory"
)

// BlockFrame stores runtime info about a control structure for branching
type BlockFrame struct {
	StartPos int // Position after block header (key into Blocks map)
}

type Memory struct {
	Globals        *memory.Global
	Memories       []iface.Memory
	Tables         []*memory.Table
	Stack          *memory.Stack[any]
	FuncSignatures []fnsig.Signature // indexed by function index
	TypeSignatures []fnsig.Signature // indexed by type index (for call_indirect)
}

type Context struct {
	*Memory

	Locals *memory.Stack[any]

	NumResults int
	Params     []any

	Body                *binary.Iterator
	FunctionCallRequest int
	Done                bool
	TailCall            bool
	Paused              bool

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
