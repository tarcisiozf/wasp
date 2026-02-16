package lex

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

func (lexer *Lexer) byte() byte {
	b := lexer.data[lexer.pos]
	lexer.pos++
	return b
}

func (lexer *Lexer) Word() string {
	var bytes []byte
	for lexer.HasNext() {
		b := lexer.Current()
		if IsBlank(b) {
			break
		}
		bytes = append(bytes, lexer.byte())
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
		bytes = append(bytes, lexer.byte())
	}

	return bytes
}

func (lexer *Lexer) Pos() int {
	return lexer.pos
}

func IsAlpha(b byte) bool {
	return b >= 'a' && b <= 'z'
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

func NewLexer(data []byte) *Lexer {
	return &Lexer{
		data: data,
		size: len(data),
		pos:  0,
	}
}
