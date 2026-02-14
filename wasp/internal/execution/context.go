package execution

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/memory"
)

// BlockKind identifies the type of control structure
type BlockKind int

const (
	BlockKindBlock BlockKind = iota
	BlockKindLoop
	BlockKindIf
)

// BlockInfo stores information about a control structure for branching
type BlockInfo struct {
	Kind      BlockKind
	StartPos  int // Position after block header (for loops)
	BlockType byte
}

type Context struct {
	Stack    *memory.Stack[any]
	Locals   *memory.Stack[any]
	Globals  *memory.Global
	Memories []*memory.Memory

	Body *binary.Iterator

	FunctionCallRequest int
	Done                bool
	Condition           bool

	// BlockStack tracks nested blocks for branching
	BlockStack *memory.Stack[BlockInfo]
}
