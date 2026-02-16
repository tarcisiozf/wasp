package ast

type Node interface {
	Position() int
	Elem() string
	ElemType() string
	Children() []Node
	String() string
}

type BaseNode struct {
	Pos int
}

var _ Node = (*BaseNode)(nil)

func (node *BaseNode) Position() int {
	return node.Pos
}

func (node *BaseNode) Elem() string {
	panic("elem not supported")
}

func (node *BaseNode) Children() []Node {
	panic("children not supported")
}

func (node *BaseNode) ElemType() string {
	panic("children not supported")
}

func (node *BaseNode) String() string {
	panic("string not supported")
}

type Comment struct {
	BaseNode
	Comment string
}

var _ Node = (*Comment)(nil)

func (c *Comment) Elem() string {
	return c.Comment
}

type EndComment Comment

var _ Node = (*EndComment)(nil)

func (e *EndComment) Elem() string {
	return e.Comment
}

type List struct {
	BaseNode
	elemType string
	children []Node
}

var _ Node = (*List)(nil)

func (list *List) Children() []Node {
	return list.children
}

type Keyword struct {
	BaseNode
	Keyword string
}

var _ Node = (*Keyword)(nil)

func (k *Keyword) Elem() string {
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
