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
	case *List, *Module, *Func:
		writeListLike(sb, depth, node)
	case *Keyword:
		sb.WriteString(node.(*Keyword).Keyword)
	case *Comment:
		sb.WriteString("(;")
		sb.WriteString(node.(*Comment).Comment)
		sb.WriteString(";) ")
	case *EndComment:
		sb.WriteString(";; ")
		sb.WriteString(node.(*EndComment).Comment)
		sb.WriteByte('\n')
	case *StringLiteral:
		sb.WriteByte('"')
		sb.WriteString(node.(*StringLiteral).Literal)
		sb.WriteByte('"')
	default:
		panic(fmt.Sprintf("not implemented: %T", node))
	}
}

func writeListLike(sb *strings.Builder, depth int, node Node) {
	sb.WriteByte('(')
	for _, child := range node.Children() {
		s(sb, depth+1, child)
		sb.WriteByte(' ')
	}
	sb.WriteByte(')')
}
