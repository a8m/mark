package mark

import (
	"fmt"
)

type Tree struct {
	text string
	lex  *lexer
	// Parsing only
	token     [3]item // three-token lookahead for parser
	peekCount int
	Nodes     []Node
}

// Parse convert the raw text to NodeTree.
func (t *Tree) parse() {
Loop:
	for {
		switch p := t.peek().typ; p {
		case eof:
			break Loop
		case itemText, itemStrong, itemItalic, itemStrike, itemCode:
			t.parseParagraph()
		default:
			fmt.Println("Nothing to do")
		}
	}
}

// next returns the next token
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// parseParagraph scan until itemBr occur.
func (t *Tree) parseParagraph() {
	// New paragraph node

	// for loop
	// if eof;
	// if br; append to tree
	// always text... and push emphasis to it.
}
