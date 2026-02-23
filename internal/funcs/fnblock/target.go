package fnblock

// Kind identifies the type of control structure
type Kind int

const (
	KindBlock Kind = iota
	KindLoop
	KindIf
)

// Target stores precomputed block information for efficient branching
type Target struct {
	Kind      Kind
	StartPos  int // Position after block header (for loops to jump back)
	ElsePos   int // Position of else branch (for if fnblock, 0 if no else)
	EndPos    int // Position after the end opcode
	BlockType byte
}
