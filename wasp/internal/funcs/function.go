package funcs

type Function struct {
	Signature Signature
	Body      []byte

	Index  int
	Offset int
}
