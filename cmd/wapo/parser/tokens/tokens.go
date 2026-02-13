package tokens

const (
	Semicolon  = ';'
	Newline    = '\n'
	OpenParen  = '('
	CloseParen = ')'
	Quote      = '"'
	Backslash  = '\\'
)

func IsBlank(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func IsWordChar(ch byte) bool {
	return ch >= 'a' && ch <= 'z'
}
