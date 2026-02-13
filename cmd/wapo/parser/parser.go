package parser

import (
	"fmt"
	"iter"
	"wasp/cmd/wapo/ast"
	"wasp/cmd/wapo/lex"
	"wasp/cmd/wapo/parser/tokens"
)

func Parse(it *lex.Iterator) iter.Seq2[ast.Node, error] {
	return func(yield func(ast.Node, error) bool) {
		for it.HasNext() {
			ch := it.Peek()

			switch ch {
			case tokens.Semicolon:
				it.SkipLine()
			case tokens.OpenParen:
				yield(parseList(it))
			default:
				it.Next()
			}
		}
	}
}

func parseList(it *lex.Iterator) (*ast.List, error) {
	depth := 0
	inString := false
	start := it.Position()
	end := -1

	for it.HasNext() && end == -1 {
		switch it.Peek() {
		case tokens.Semicolon:
			if !inString {
				readComment(it)
			}
		case tokens.Quote:
			if it.PeekPrev() != tokens.Backslash {
				inString = !inString
			}
		case tokens.OpenParen:
			if !inString {
				depth++
			}
		case tokens.CloseParen:
			if !inString {
				depth--
				if depth == 0 {
					end = it.Position() + 1
				}
			}
		}
		it.Next()
	}
	if end == -1 {
		return nil, fmt.Errorf("unexpected end of list, line: %d, col: %d", it.Line(start), it.Col(start))
	}
	content := string(it.Range(start, end))
	return ast.NewList(content, start), nil
}

func readComment(it *lex.Iterator) {
	if it.PeekNext() == tokens.Semicolon {
		it.SkipLine()
	} else if it.PeekPrev() == tokens.OpenParen {
		it.ReadUntil(tokens.Semicolon)
	}
}
