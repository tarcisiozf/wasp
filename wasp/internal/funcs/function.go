package funcs

// BlockKind identifies the type of control structure
type BlockKind int

const (
	BlockKindBlock BlockKind = iota
	BlockKindLoop
	BlockKindIf
)

// BlockTarget stores precomputed block information for efficient branching
type BlockTarget struct {
	Kind      BlockKind
	StartPos  int // Position after block header (for loops to jump back)
	ElsePos   int // Position of else branch (for if blocks, 0 if no else)
	EndPos    int // Position after the end opcode
	BlockType byte
}

type Function struct {
	Signature Signature
	Locals    []any
	Body      []byte

	// Blocks maps block start positions to their precomputed targets
	Blocks map[int]BlockTarget

	Index  int
	Offset int
}
