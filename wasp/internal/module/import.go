package module

import (
	"fmt"
	"wasp/wasp/internal/funcs"
)

type Import struct {
	ModuleName string
	FieldName  string
	Kind       byte
	Signature  funcs.Signature
}

func (i Import) String() string {
	str := i.ModuleName + "." + i.FieldName
	switch i.Kind {
	case exportKindFunc:
		str += " " + i.Signature.String()
	default:
		str += fmt.Sprintf(" (unknown kind 0x%x)", i.Kind)
	}
	return str
}
