package lex

import (
	"fmt"
	"wasp/cmd/wat/wat/tokens"
)

type Lexer struct {
	data []byte
	size int
	pos  int
}

func (lexer *Lexer) HasNext() bool {
	return lexer.pos < lexer.size
}

func (lexer *Lexer) Peek() byte {
	return lexer.data[lexer.pos]
}

func (lexer *Lexer) Next() byte {
	b := lexer.data[lexer.pos]
	lexer.pos++
	return b
}

func (lexer *Lexer) Assert(expected ...byte) error {
	has := lexer.Bytes(len(expected))
	for i, b := range expected {
		if has[i] != b {
			return fmt.Errorf("expected %q at position %d, got %q", b, lexer.pos-len(expected)+i, has[i])
		}
	}
	return nil
}

func (lexer *Lexer) Bytes(n int) []byte {
	data := lexer.data[lexer.pos : lexer.pos+n]
	lexer.pos += n
	return data
}

func (lexer *Lexer) Skip() {
	for isEmpty(lexer.Peek()) {
		lexer.Next()
	}
}

func (lexer *Lexer) ReadUntil(delimiter byte) []byte {
	var data []byte
	for lexer.HasNext() {
		ch := lexer.Next()
		if ch == delimiter {
			break
		}
		data = append(data, ch)
	}
	return data
}

func (lexer *Lexer) Position() int {
	return lexer.pos
}

func (lexer *Lexer) Keyword() (string, error) {
	var keyword []byte
	for lexer.HasNext() {
		ch := lexer.Peek()
		if isEndOfSequence(ch) {
			break
		}
		if ch > 'a' && ch < 'z' || ch > 'A' && ch < 'Z' {
			keyword = append(keyword, lexer.Next())
		} else {
			return "", fmt.Errorf("unexpected character %q at position %d", ch, lexer.pos)
		}
	}
	return string(keyword), nil
}

func isEmpty(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func isEndOfSequence(ch byte) bool {
	return isEmpty(ch) || tokens.IsToken(ch)
}

func NewLexer(data []byte) *Lexer {
	return &Lexer{
		data: data,
		size: len(data),
	}
}
