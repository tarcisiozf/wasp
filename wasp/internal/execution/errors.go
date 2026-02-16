package execution

import "errors"

var (
	ErrUnreachable              = errors.New("unreachable")
	ErrInvalidTableIndex        = errors.New("invalid table index")
	ErrUndefinedElement         = errors.New("undefined element")
	ErrUninitializedElement     = errors.New("uninitialized element")
	ErrIndirectCallTypeMismatch = errors.New("indirect call type mismatch")
)
