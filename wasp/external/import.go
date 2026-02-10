package external

import "wasp/wasp/funcs"

type Import struct {
	ModuleName string
	FieldName  string
	Kind       byte
	Signature  funcs.Signature
}
