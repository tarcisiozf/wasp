package parser

import (
	"fmt"
	"iter"
	"wasp/cmd/wat/wat/ast"
	"wasp/cmd/wat/wat/lex"
	"wasp/cmd/wat/wat/tokens"
)

func Parse(lexer *lex.Lexer) (ast.Node, error) {
	program := ast.Program{}
	for node, err := range parse(lexer) {
		if err != nil {
			return nil, err
		}
		program.Append(node)
	}
	return program, nil
}

func parse(lexer *lex.Lexer) iter.Seq2[ast.Node, error] {
	return func(yield func(ast.Node, error) bool) {
		for lexer.HasNext() {
			switch lexer.Next() {
			case tokens.Semicolon:
				yield(parseComment(lexer))
			case tokens.OpenParen:
				yield(parseList(lexer))
			}
		}
	}
}

func parseList(lexer *lex.Lexer) (ast.Node, error) {
	lexer.Skip()
	keyword, err := lexer.Keyword()
	if err != nil {
		return nil, fmt.Errorf("expected keyword at position %d", lexer.Position())
	}
	switch keyword {
	case "module":
		return parseModule(lexer)
	}
	return nil, nil
}

func parseModule(lexer *lex.Lexer) (ast.Node, error) {
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
