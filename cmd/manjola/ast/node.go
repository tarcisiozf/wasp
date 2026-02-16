package ast

import "fmt"

type Node interface {
	Position() int
	Children() []Node
	String() string
}

type BaseNode struct {
	pos      int
	children []Node
}

func (node *BaseNode) Children() []Node {
	return node.children
}

var _ Node = (*BaseNode)(nil)

func (node *BaseNode) Position() int {
	return node.pos
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
}

var _ Node = (*List)(nil)

func (list *List) String() string {
	return fmt.Sprintf("List<%T>(%d)", list, len(list.children))
}

type Keyword struct {
	BaseNode
	Keyword string
}

var _ Node = (*Keyword)(nil)

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

type Module List

type Func List

type ListLike interface {
	Module | Func
}
