package mark

type Tree struct {
	text string
	lex  *lexer
	// Parsing only
	token     [3]item // three-token lookahead for parser
	peekCount int
}

// next returns the next token
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[peekCount]
}
