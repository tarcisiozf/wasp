package main

type List struct {
	nodeType string
	content  string
	offset   int
}

func NewList(content string, offset int) *List {
	return &List{
		nodeType: readNodeType(content),
		content:  content,
		offset:   offset,
	}
}

func readNodeType(data string) string {
	for i := 1; i < len(data); i++ {
		if isWordChar(data[i]) {
			continue
		}
		return data[1 : i-1]
	}
	panic("invalid")
}
