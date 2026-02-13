package ast

import (
	"regexp"
	"wasp/cmd/wapo/types"
)

var keywordPattern = regexp.MustCompile(`[^@](\w+)`)

type List struct {
	nodeType types.Type
	content  string
	offset   int
}

func NewList(content string, offset int) *List {
	return &List{
		nodeType: types.MustParse(readNodeType(content)),
		content:  content,
		offset:   offset,
	}
}

func (list *List) String() string {
	return list.content
}

func readNodeType(data string) string {
	matches := keywordPattern.FindStringSubmatch(data)
	if len(matches) == 0 {
		return "unknown"
	}
	return matches[1]
}
