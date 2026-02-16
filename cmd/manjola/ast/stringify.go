package ast

import (
	"fmt"
	"strings"
)

func Stringify(node Node) string {
	var sb strings.Builder
	s(&sb, node)
	return sb.String()
}

func s(sb *strings.Builder, node Node) {
	switch node.(type) {
	case *List:
		sb.WriteByte('(')
		for _, child := range node.Children() {
			s(sb, child)
			sb.WriteByte(' ')
		}
		sb.WriteByte(')')
	case *Keyword:
		sb.WriteString(node.Elem())
	case *Comment:
		sb.WriteString("(;")
		sb.WriteString(node.Elem())
		sb.WriteString(";) ")
	case *EndComment:
		sb.WriteString(";; ")
		sb.WriteString(node.Elem())
		sb.WriteByte('\n')
	default:
		panic(fmt.Sprintf("not implemented: %T", node))
	}
}
