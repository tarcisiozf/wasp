package ast

type Node interface {
	Position() int
	Elem() string
	Children() []Node
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
	//TODO implement me
	panic("children not supported")
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
