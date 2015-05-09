package mark

import (
	"fmt"
	"unicode/utf8"
)

// type position
type Pos int

// itemType identifies the type of lex items.
type itemType int

// Item represent a token or text string returned from the scanner
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

const eof = -1 // Zero value so closed channel delivers EOF

const (
	itemError itemType = iota // Error occurred; value is text of error
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
	itemItalic
	itemStrike
	itemCode
	itemImages
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
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
	close(l.items)
}

// next return the next rune in the input
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// lexAny scans non-space items.
func lexAny(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		return nil
	default:
		fmt.Printf("unrecognized character: %#U\n", r)
		return lexAny
	}
}
