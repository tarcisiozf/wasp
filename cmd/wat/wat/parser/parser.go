package parser

import (
	"fmt"
	"iter"
	"wasp/cmd/wat/wat/ast"
	"wasp/cmd/wat/wat/lex"
	"wasp/cmd/wat/wat/syntax"
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
			ch := lexer.Next()
			switch ch {
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
	keyword, err := parseKeyword(lexer)
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
	lexer.Skip()
	keywords, err := readKeywords(lexer)
	if err != nil {
		return nil, err
	}

	if len(keywords) == 1 {
		if keywords[0] == syntax.Binary {
			return parseBinaryModule(lexer)
		}
	} else if len(keywords) > 1 {
		return nil, fmt.Errorf("expected only one keyword at position %d", lexer.Position())
	}

	return nil, nil
}

func parseBinaryModule(lexer *lex.Lexer) (ast.BinaryModule, error) {
	list, err := readStrings(lexer)
	if err != nil {
		return ast.BinaryModule{}, err
	}
	var blob []byte
	for _, encoded := range list {
		decoded := decodeString(encoded)
		blob = append(blob, decoded...)
	}
	return ast.BinaryModule{
		Blob: blob,
	}, nil
}

func decodeString(s string) []byte {
	off := 0
	size := len(s)
	var data []byte
	for off < size {
		ch := s[off]
		b := ch
		if ch == tokens.Escape {
			if off+3 > size {
				break
			}
			b = (hexToByte(s[off+1]) << 4) | hexToByte(s[off+2])
			off += 2
		}
		data = append(data, b)
		off++
	}
	return data
}

func hexToByte(u uint8) uint8 {
	if u >= '0' && u <= '9' {
		return u - '0'
	} else if u >= 'a' && u <= 'f' {
		return u - 'a' + 10
	} else if u >= 'A' && u <= 'F' {
		return u - 'A' + 10
	}
	panic(fmt.Sprintf("invalid hex character: %q", u))
}

func readStrings(lexer *lex.Lexer) ([]string, error) {
	var list []string
	for lexer.HasNext() {
		lexer.Skip()
		ch := lexer.Peek()
		if ch == tokens.DoubleQuote {
			str, err := parseString(lexer)
			if err != nil {
				return nil, err
			}
			list = append(list, str)
		} else if ch == tokens.Semicolon {
			lexer.JumpLine()
		} else {
			break
		}
	}
	return list, nil
}

func parseString(lexer *lex.Lexer) (string, error) {
	if err := lexer.Assert(tokens.DoubleQuote); err != nil {
		return "", fmt.Errorf("expected '\"' at position %d", lexer.Position())
	}
	var str []byte
	for lexer.HasNext() {
		ch := lexer.Next()
		if ch == tokens.DoubleQuote && lexer.PeekAt(-1) != tokens.Escape {
			break
		}
		str = append(str, ch)
	}
	return string(str), nil
}

func readKeywords(lexer *lex.Lexer) ([]syntax.Keyword, error) {
	var list []syntax.Keyword
	for lexer.HasNext() {
		lexer.Skip()
		ch := lexer.Peek()
		if lex.IsAlpha(ch) {
			keyword, err := parseKeyword(lexer)
			if err != nil {
				return list, err
			}
			list = append(list, keyword)
		} else {
			break
		}
	}
	return list, nil
}

func parseKeyword(lexer *lex.Lexer) (syntax.Keyword, error) {
	var keyword []byte
	for lexer.HasNext() {
		ch := lexer.Peek()
		if lex.IsEndOfSequence(ch) {
			break
		}
		if lex.IsAlpha(ch) {
			keyword = append(keyword, lexer.Next())
		} else {
			return "", fmt.Errorf("unexpected character %q at position %d", ch, lexer.Position())
		}
	}
	return syntax.MustParse(string(keyword)), nil
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
