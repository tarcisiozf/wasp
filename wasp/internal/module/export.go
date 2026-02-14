package module

type Export struct {
	kind  int
	index int
}

func (e Export) Kind() int {
	return e.kind
}
