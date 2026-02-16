package ast

import "fmt"

type Node interface {
	Position() int
	String() string
}

type BaseNode struct {
	Pos int
}

var _ Node = (*BaseNode)(nil)

func (node *BaseNode) Position() int {
	return node.Pos
}

func (node *BaseNode) String() string {
	panic("String() not implemented")
}

type Comment struct {
	BaseNode
	Comment string
}

var _ Node = (*Comment)(nil)

type EndComment Comment

var _ Node = (*EndComment)(nil)

type List struct {
	BaseNode
	ElemType string
	Children []Node
}

var _ Node = (*List)(nil)

func (list *List) String() string {
	return fmt.Sprintf("List<%s>(%d)", list.ElemType, len(list.Children))
}

type Keyword struct {
	BaseNode
	Keyword string
}

var _ Node = (*Keyword)(nil)

func (k *Keyword) Elem() string {
	return k.Keyword
}

func (k *Keyword) String() string {
	return k.Keyword
}

type StringLiteral struct {
	BaseNode
	Literal string
}

var _ Node = (*StringLiteral)(nil)

func (s *StringLiteral) Elem() string {
	return s.Literal
}
