package lex

import (
	"fmt"
	"wasp/cmd/wapo/tokens"
)

type Lexer struct {
	data []byte
	size int
	pos  int
}

func (lexer *Lexer) HasNext() bool {
	return lexer.pos < len(lexer.data)
}

func (lexer *Lexer) Current() byte {
	return lexer.data[lexer.pos]
}

func (lexer *Lexer) Next() {
	lexer.pos++
}

func (lexer *Lexer) Byte() byte {
	b := lexer.data[lexer.pos]
	lexer.pos++
	return b
}

func (lexer *Lexer) Word() string {
	var bytes []byte
	for lexer.HasNext() {
		b := lexer.Current()
		if !IsKeywordChar(b) {
			break
		}
		bytes = append(bytes, lexer.Byte())
	}
	return string(bytes)
}

func (lexer *Lexer) Peek() byte {
	return lexer.data[lexer.pos+1]
}

func (lexer *Lexer) ReadUntil(target byte) []byte {
	var bytes []byte
	for lexer.HasNext() {
		b := lexer.Current()
		if b == target {
			break
		}
		bytes = append(bytes, lexer.Byte())
	}

	return bytes
}

func (lexer *Lexer) Pos() int {
	return lexer.pos
}

func (lexer *Lexer) Assert(expected ...byte) {
	offset := lexer.Pos()
	numBytes := len(expected)

	if offset+numBytes > lexer.size {
		fmt.Printf(
			"unexpected end of input at position %d, expected '%s' got '%s'\n",
			offset,
			string(expected),
			string(lexer.Range(offset, offset+numBytes)),
		)
		fmt.Println(lexer.String())
		panic("assertion failed")
	}

	for i, b := range expected {
		got := lexer.Byte()
		if got == b {
			continue
		}
		fmt.Printf(
			"unexpected byte '%c' at position %d, expected '%s' got '%s' at offset %d\n",
			got, i,
			string(expected),
			string(lexer.Range(offset, offset+numBytes+10)),
			offset,
		)
		fmt.Println(lexer.String())
		panic("assertion failed")
	}
}

func (lexer *Lexer) Range(start, end int) []byte {
	return lexer.data[start:end]
}

func IsAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func IsNumeric(b byte) bool {
	return b >= '0' && b <= '9'
}

func IsAlphaNumeric(b byte) bool {
	return IsAlpha(b) || IsNumeric(b) || b == '-'
}

func IsBlank(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func IsVar(b byte) bool {
	return b == '$'
}

func IsComment(b byte) bool {
	return b == ';'
}

func IsKeywordChar(b byte) bool {
	return IsAlphaNumeric(b) ||
		IsVar(b) ||
		b == '.' ||
		b == '_' ||
		b == '|' ||
		b == ':' ||
		b == '=' ||
		b == '+'
}

func (lexer *Lexer) Line(index int) int {
	var line int
	for i := 0; i < index; i++ {
		if lexer.data[i] == tokens.Newline {
			line++
		}
	}
	return line + 1
}

func (lexer *Lexer) Col(index int) int {
	var line int
	for i := 0; i < index; i++ {
		if lexer.data[i] == tokens.Newline {
			line = i
		}
	}
	return index - line
}

func (lexer *Lexer) String() string {
	str := fmt.Sprintf("line: %d, col: %d\n", lexer.Line(lexer.pos), lexer.Col(lexer.pos))
	start := lexer.pos - 20
	if start < 0 {
		start = 0
	}
	end := lexer.pos + 10
	if end > lexer.size {
		end = lexer.size
	}
	line := make([]byte, 0, end-start)
	for i := start; i < end; i++ {
		if lexer.data[i] == tokens.Newline {
			line = append(line, '\\', 'n')
		} else {
			line = append(line, lexer.data[i])
		}
	}
	str += string(line) + "\n"
	str += fmt.Sprintf("%s^", string(make([]byte, lexer.pos-start+1)))
	str += "\n-----------------\n"
	return str
}

func (lexer *Lexer) Prev() byte {
	return lexer.data[lexer.pos-1]
}

func NewLexer(data []byte) *Lexer {
	return &Lexer{
		data: data,
		size: len(data),
		pos:  0,
	}
}
