package module

import (
	"wasp/wasp/internal/funcs"
)

type Import struct {
	ModuleName string
	FieldName  string
	Kind       byte
	Signature  funcs.Signature
}
