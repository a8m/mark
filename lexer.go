package mark

import (
	"fmt"
	"regexp"
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
	itemGfmCodeBlock
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

// Block Grammer
var block = map[string]*regexp.Regexp{
	"heading": regexp.MustCompile("^ *(#{1,6}) *([^\n]+?) *#* *(?:\n+|$)"),
	"hr":      regexp.MustCompile("^( *[-*_]){3,} *(?:\n+|$)"),
	"code":    regexp.MustCompile("^( {4}[^\n]+\n*)+"),
	// Backreferences is unavailable
	// TODO(Ariel): it's ugly, remove this duplicate
	"gfm-code-1": regexp.MustCompile("^`{3,} *(\\S+)? *\n([\\s\\S]+?)\\s*`{3,}$*(?:\n+|$)"),
	"gfm-code-2": regexp.MustCompile("^~{3,} *(\\S+)? *\n([\\s\\S]+?)\\s*~{3,}$*(?:\n+|$)"),
}

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
	switch r := l.next(); r {
	case eof:
		return nil
	case '*', '-':
		l.backup()
		return lexHr
	case '#':
		l.backup()
		return lexHeading
	case ' ':
		// Should be here ?
		if block["code"].MatchString(l.input[l.pos-1:]) {
			l.backup()
			return lexCode
		}
		fallthrough
	case '`', '~':
		// if it's gfm-code
		c := l.input[l.pos : l.pos+2]
		if c == "``" || c == "~~" {
			l.backup()
			return lexGfmCode
		}
		fallthrough
	default:
		l.backup()
		return lexText
	}
}

// lexHeading scans heading items.
func lexHeading(l *lexer) stateFn {
	if block["heading"].MatchString(l.input[l.pos:]) {
		match := block["heading"].FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemHeading)
		return lexAny
	}
	return lexText
}

// lexHr scans horizontal rules items.
func lexHr(l *lexer) stateFn {
	if block["hr"].MatchString(l.input[l.pos:]) {
		match := block["hr"].FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemHr)
		return lexAny
	}
	return lexText
}

// lexGfmCode scans GFM code block.
func lexGfmCode(l *lexer) stateFn {
	fmt.Println("Inside gfm")
	re := block["gfm-code-1"]
	// if it's the ~ version
	if l.peek() == '~' {
		re = block["gfm-code-2"]
	}
	if re.MatchString(l.input[l.pos:]) {
		match := re.FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemGfmCodeBlock)
		return lexAny
	}
	return lexText
}

// lexCode scans code block.
func lexCode(l *lexer) stateFn {
	match := block["code"].FindString(l.input[l.pos:])
	l.pos += Pos(len(match))
	l.emit(itemCodeBlock)
	return lexAny
}

// lexText scans until eol(\n)
func lexText(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case r == '\n' && l.peek() == '\n' || r == ' ' && l.peek() == ' ':
			// if we inside a paragraph(drain it before we emit new-line)
			if l.pos > l.start {
				l.emit(itemParagraph)
			}
			// length of new-line
			l.pos += Pos(1)
			l.emit(itemNewLine)
			break Loop
		// if it's start as an emphasis
		case r == '`', r == '_', r == '~', r == '*':
			// test with regex which of them(if not, fallthrough)
		case r == eof:
			fmt.Println("end of file")
			break Loop
		default:
			// mm..
		}
	}
	return lexAny
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}
