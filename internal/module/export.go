package module

import "fmt"

type Export struct {
	name  string
	kind  byte
	index int
}

func (e Export) String() string {
	return fmt.Sprintf("%s (kind: 0x%x)", e.name, e.kind)
}
