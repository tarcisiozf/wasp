package types

type Type string

var types = make(map[string]Type)

var (
	Module           = addType("module")
	AssertMalformed  = addType("assert_malformed")
	AssertInvalid    = addType("assert_invalid")
	AssertReturn     = addType("assert_return")
	AssertTrap       = addType("assert_trap")
	AssertExhaustion = addType("assert_exhaustion")
)

func addType(name string) Type {
	t := Type(name)
	types[name] = t
	return t
}

func MustParse(s string) Type {
	if t, ok := types[s]; ok {
		return t
	}
	panic("unknown type: " + s)
}
