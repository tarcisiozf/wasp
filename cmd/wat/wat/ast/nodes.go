package ast

type Node interface {
}

type Program struct {
	Body []Node
}

func (f *Program) Append(node Node) {
	f.Body = append(f.Body, node)
}

type LeadingComment struct {
	Value string
}

type Module struct {
	Fields []Node
}

type BinaryModule struct {
	Blob []byte
}
