package parser

import (
	"fmt"
	"wasp/cmd/wat/wat/ast"
	"wasp/cmd/wat/wat/tokens"
)

func Parse(data []byte) (ast.Node, error) {
	lexer := NewLexer(data)
	program := ast.Program{}
	for lexer.HasNext() {
		switch lexer.Next() {
		case tokens.Semicolon:
			comment, err := parseComment(lexer)
			if err != nil {
				return nil, err
			}
			program.Append(comment)
		case tokens.OpenParen:
			node, err := parseList(lexer)
			if err != nil {
				return nil, err
			}
			program.Append(node)
		}
	}
	return program, nil
}

func parseList(lexer *Lexer) (ast.Node, error) {
	return nil, nil
}

func parseComment(lexer *Lexer) (ast.LeadingComment, error) {
	if err := lexer.Assert(tokens.Semicolon); err != nil {
		return ast.LeadingComment{}, fmt.Errorf("expected ';;' at position %d", lexer.pos)
	}
	lexer.Skip()
	value := lexer.ReadUntil(tokens.NewLine)
	return ast.LeadingComment{
		Value: string(value),
	}, nil
}
