package funcs

import (
	"strings"
	"wasp/wasp/internal/types"
)

type Signature struct {
	Params  []byte
	Results []byte
}

func (s Signature) String() string {
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = types.ForCode(p).String()
	}
	results := make([]string, len(s.Results))
	for i, r := range s.Results {
		results[i] = types.ForCode(r).String()
	}
	str := "func(" + strings.Join(params, ", ") + ")"
	if len(results) > 0 {
		str += " -> (" + strings.Join(results, ", ") + ")"
	}
	return str
}
