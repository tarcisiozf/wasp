package tokens

var isToken = make([]bool, 256)

func addToken(ch byte) byte {
	isToken[ch] = true
	return ch
}

var (
	Semicolon  = addToken(';')
	NewLine    = addToken('\n')
	OpenParen  = addToken('(')
	CloseParen = addToken(')')
)

func IsToken(ch byte) bool {
	return isToken[ch]
}
