package funcs

type Function struct {
	Signature Signature
	Locals    []any
	Body      []byte

	Index  int
	Offset int
}
