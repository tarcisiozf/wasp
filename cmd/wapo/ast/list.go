package ast

import "regexp"

var keywordPattern = regexp.MustCompile(`(\w+)`)

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

func (list *List) String() string {
	return list.content
}

func readNodeType(data string) string {
	keyword := keywordPattern.FindString(data)
	if keyword == "" {
		return "unknown"
	}
	return keyword
}
