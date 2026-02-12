package syntax

type Keyword string

var keywords = map[string]Keyword{}

func addKeyword(s string) Keyword {
	k := Keyword(s)
	keywords[s] = k
	return k
}

var (
	Binary = addKeyword("binary")
	Module = addKeyword("module")
)

func MustParse(s string) Keyword {
	if k, ok := keywords[s]; ok {
		return k
	}
	panic("unknown keyword: " + s)
}
