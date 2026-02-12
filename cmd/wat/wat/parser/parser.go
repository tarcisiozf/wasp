package parser

import (
	"fmt"
	"wasp/cmd/wat/wat/ast"
	"wasp/cmd/wat/wat/lex"
	"wasp/cmd/wat/wat/tokens"
)

func Parse(lexer *lex.Lexer) (ast.Node, error) {
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

func parseList(lexer *lex.Lexer) (ast.Node, error) {
	return nil, nil
}

func parseComment(lexer *lex.Lexer) (ast.LeadingComment, error) {
	if err := lexer.Assert(tokens.Semicolon); err != nil {
		return ast.LeadingComment{}, fmt.Errorf("expected ';;' at position %d", lexer.Position())
	}
	lexer.Skip()
	value := lexer.ReadUntil(tokens.NewLine)
	return ast.LeadingComment{
		Value: string(value),
	}, nil
}
