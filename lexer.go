package mark

import (
	"go/token"
	"unicode/utf8"
)

// itemType identifies the type of lex items.
type itemType int

// Item represent a token or text string returned from the scanner
type item struct {
	typ itemType  // The type of this item.
	pos token.Pos // The starting position, in bytes, of this item in the input string.
	val string    // The value of this item.
}

const (
	itemEOF   itemType = iota - 1 // Zero value so closed channel delivers EOF
	itemError                     // Error occurred; value is text of error
	// Intersting things
	itemNewLine
	itemHTML
	// Block Elements
	itemParagraph
	itemLineBreak
	itemHeading
	itemLHeading
	itemBlockQuote
	itemList
	itemCodeBlock
	itemHr
	itemTable
	// Span Elements
	itemLinks
	itemEmphasis
	itemCode
	itemImages
)

// Regexp grammer
const (
	code      = "/^`/"
	codeBlock = "/^```/"
	heading   = "/^#/"
	lheading  = "/^([^\n]+)\n *(=|-){2,} *(?:\n+|$)/"
	comment   = "/<!--[\\s\\S]*?-->/"
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        token.Pos // current position in the input
	start      token.Pos // start position of this item
	width      token.Pos // width of last rune read from input
	lastPos    token.Pos // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// lex creates a new lexer for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
}

// next return the next rune in the input
func (l *lexer) next() rune {
	if !l.done && int(l.pos) >= len(l.input) {
		l.width = 0
		return itemEOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = token.Pos(w)
	l.pos += l.width
	return r
}

// lexAny scans non-space items.
func lexAny(l *lexer) stateFn {
	switch r := l.next(); {
	default:
		fmt.Println("unrecognized character: %#U", r)
		return lexAny
	}
}
