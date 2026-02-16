package ast

import (
	"fmt"
	"strings"
)

func Stringify(node Node) string {
	var sb strings.Builder
	s(&sb, 0, node)
	return sb.String()
}

func s(sb *strings.Builder, depth int, node Node) {
	switch node.(type) {
	case *List:
		list := node.(*List)
		elemType := list.ElemType()

		sb.WriteByte('(')
		if elemType != "" {
			sb.WriteString(elemType)
			sb.WriteByte('\n')
		}
		for _, child := range list.Children() {
			s(sb, depth+1, child)
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
	case *StringLiteral:
		sb.WriteByte('"')
		sb.WriteString(node.Elem())
		sb.WriteByte('"')
	default:
		panic(fmt.Sprintf("not implemented: %T", node))
	}
}
