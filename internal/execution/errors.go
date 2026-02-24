package execution

import "errors"

var (
	ErrIntegerOverflow            = errors.New("integer overflow")
	ErrInvalidConversionToInteger = errors.New("invalid conversion to integer")
	ErrUnreachable                = errors.New("unreachable")
	ErrInvalidTableIndex          = errors.New("invalid table index")
	ErrUndefinedElement           = errors.New("undefined element")
	ErrUninitializedElement       = errors.New("uninitialized element")
	ErrIndirectCallTypeMismatch   = errors.New("indirect call type mismatch")
	ErrInvalidSignatureIndex      = errors.New("invalid signature index")
	ErrInvalidFunctionIndex       = errors.New("invalid function index")
)
