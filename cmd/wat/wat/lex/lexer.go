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
	return lexer.PeekAt(0)
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
	for lexer.HasNext() {
		ch := lexer.Peek()
		if IsBlank(ch) || ch == tokens.NewLine {
			lexer.Next()
		} else {
			break
		}
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

func (lexer *Lexer) PeekAt(at int) byte {
	return lexer.data[lexer.pos+at]
}

func (lexer *Lexer) JumpLine() {
	lexer.ReadUntil('\n')
}

func IsAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func IsBlank(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func IsEndOfSequence(ch byte) bool {
	return IsBlank(ch) || tokens.IsToken(ch)
}

func NewLexer(data []byte) *Lexer {
	return &Lexer{
		data: data,
		size: len(data),
	}
}
