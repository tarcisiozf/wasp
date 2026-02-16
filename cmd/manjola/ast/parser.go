package ast

import (
	"fmt"
	"iter"
	"wasp/cmd/manjola/lex"
)

func Parse(lexer *lex.Lexer) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for lexer.HasNext() {
			node := parse(lexer)
			if node == nil {
				return
			}

			if !yield(node) {
				return
			}
		}
	}
}

func parse(lexer *lex.Lexer) Node {
	for lexer.HasNext() {
		b := lexer.Current()

		if b == '(' {
			return parseList(lexer)
		} else if lex.IsKeywordChar(b) {
			return parseKeyword(lexer)
		} else if lex.IsBlank(b) {
			lexer.Next()
			continue
		} else if lex.IsComment(b) {
			return parseComment(lexer)
		} else if b == '"' {
			return parseString(lexer)
		}

		fmt.Printf("Unexpected byte '%c' at position %d\n", b, lexer.Pos())
		fmt.Println(lexer.String())
		panic("unreachable")
	}
	return nil
}

func parseString(lexer *lex.Lexer) Node {
	lexer.Assert('"')

	var literal []byte
	for lexer.HasNext() {
		b := lexer.Current()
		if b == '"' && lexer.Prev() != '\\' {
			break
		}
		literal = append(literal, lexer.Byte())
	}

	lexer.Assert('"')

	return &StringLiteral{
		BaseNode{lexer.Pos()},
		string(literal),
	}
}

func parseComment(lexer *lex.Lexer) Node {
	lexer.Assert(';')

	if lexer.Current() == ';' {
		return &EndComment{
			BaseNode{lexer.Pos()},
			string(lexer.ReadUntil('\n')),
		}
	}

	comment := &Comment{
		BaseNode{lexer.Pos()},
		string(lexer.ReadUntil(';')),
	}

	lexer.Assert(';', ')')

	return comment
}

func parseKeyword(lexer *lex.Lexer) Node {
	pos := lexer.Pos()
	keyword := lexer.Word()
	return &Keyword{
		BaseNode{pos},
		keyword,
	}
}

func parseList(lexer *lex.Lexer) Node {
	lexer.Assert('(')

	pos := lexer.Pos()

	var children []Node
	for lexer.HasNext() {
		b := lexer.Current()
		if b == ')' {
			break
		}

		child := parse(lexer)
		if child == nil {
			break
		}
		children = append(children, child)
	}

	if lexer.HasNext() {
		lexer.Assert(')')
	}

	numChildren := len(children)
	var elemType string
	if numChildren > 0 {
		first := children[0]

		switch first.(type) {
		case *Comment:
			if numChildren == 1 {
				return first
			}
		case *Keyword:
			kw := first.(*Keyword)
			if isListType(kw.Elem()) {
				elemType = kw.Elem()
				children = children[1:]
			}
		}
	}

	return &List{
		BaseNode{pos},
		elemType,
		children,
	}
}

func isListType(str string) bool {
	switch str {
	case "module", "param", "result", "func",
		"type", "import", "local", "mut":
		return true
	}
	return false
}
