package ast

import (
	"fmt"
	"iter"
	"wasp/cmd/manjola/lex"
)

func Parse(lexer *lex.Lexer) []Node {
	var nodes []Node
	for node := range parse(lexer) {
		nodes = append(nodes, node)
	}

	return nodes
}

func parse(lexer *lex.Lexer) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		for lexer.HasNext() {
			b := lexer.Current()

			if b == '(' {
				yield(parseList(lexer))
			} else if lex.IsAlphaNumeric(b) || lex.IsVar(b) {
				yield(parseKeyword(lexer))
			} else if lex.IsBlank(b) {
				lexer.Next()
			} else if lex.IsComment(b) {
				yield(parseComment(lexer))
			} else {
				fmt.Printf("Unexpected byte: '%c'\n", b)
				panic("unreachable")
			}
		}
	}
}

func parseComment(lexer *lex.Lexer) Node {
	if lexer.Peek() == ';' {
		return &EndComment{
			BaseNode{lexer.Pos()},
			string(lexer.ReadUntil('\n')),
		}
	}

	lexer.Next()

	return &Comment{
		BaseNode{lexer.Pos()},
		string(lexer.ReadUntil(';')),
	}
}

func parseKeyword(lexer *lex.Lexer) Node {
	return &Keyword{
		BaseNode{lexer.Pos()},
		lexer.Word(),
	}
}

func parseList(lexer *lex.Lexer) Node {
	if lexer.Current() != '(' {
		panic("Expected '(' at the beginning of a list")
	}

	pos := lexer.Pos()
	lexer.Next()

	var children []Node
	for lexer.HasNext() {
		b := lexer.Peek()
		if b == ')' {
			lexer.Next() // consume ')'
			break
		}

		for child := range parse(lexer) {
			children = append(children, child)
		}
	}

	//if len(node.children) > 0 && node.children[0].t == nodeTypeKeyword {
	//	elemType := node.children[0].elem
	//	return List{node, elemType}
	//}

	return &List{
		BaseNode{pos},
		children,
	}
}
