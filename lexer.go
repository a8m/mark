package mark

import (
	"fmt"
	"regexp"
	"strings"
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
	itemText
	itemLineBreak
	itemHeading
	itemLHeading // Setext-style headers
	itemBlockQuote
	itemList
	itemCodeBlock
	itemGfmCodeBlock
	itemHr
	itemTable
	itemLpTable
	// Span Elements
	itemLink
	itemAutoLink
	itemGfmLink
	itemStrong
	itemItalic
	itemStrike
	itemCode
	itemImage
	itemBr
	itemPipe
	// Indentation
	itemIndent
)

var (
	reEmphasise = `^_{%[1]d}([\s\S]+?(?:_{0,}))_{%[1]d}|^\*{%[1]d}([\s\S]+?(?:\*{0,}))\*{%[1]d}`
	reGfmCode   = `^%[1]s{3,} *(\S+)? *\n([\s\S]+?)\s*%[1]s{3,}$*(?:\n+|$)`
	reLinkText  = `(?:\[[^\]]*\]|[^\[\]]|\])*`
	reLinkHref  = `\s*<?([\s\S]*?)>?(?:\s+['"]([\s\S]*?)['"])?\s*`
)

// Block Grammer
var block = map[itemType]*regexp.Regexp{
	itemHeading:   regexp.MustCompile(`^ *(#{1,6}) *([^\n]+?) *#* *(?:\n+|$)`),
	itemLHeading:  regexp.MustCompile(`^([^\n]+)\n *(=|-){2,} *(?:\n+|$)`),
	itemHr:        regexp.MustCompile(`^( *[-*_]){3,} *(?:\n+|$)`),
	itemCodeBlock: regexp.MustCompile(`^(( {4}|\t)[^-+*(\d\.)\n]+\n*)+`),
	// Backreferences is unavailable
	itemGfmCodeBlock: regexp.MustCompile(fmt.Sprintf(reGfmCode, "`") + "|" + fmt.Sprintf(reGfmCode, "~")),
	// `^(?:[*+-]|\d+\.) [\s\S]+?(?:\n|)`
	itemList: regexp.MustCompile(`^(?:[*+-]|\d+\.) +?(?:\n|)`),
	// leading-pipe table
	itemLpTable: regexp.MustCompile(`^ *\|(.+)\n *\|( *[-:]+[-| :]*)\n((?: *\|.*(?:\n|$))*)\n*`),
	itemTable:   regexp.MustCompile(`^ *(\S.*\|.*)\n *([-:]+ *\|[-| :]*)\n((?:.*\|.*(?:\n|$))*)\n*`),
}

// Inline Grammer
var span = map[itemType]*regexp.Regexp{
	itemItalic: regexp.MustCompile(fmt.Sprintf(reEmphasise, 1)),
	itemStrong: regexp.MustCompile(fmt.Sprintf(reEmphasise, 2)),
	itemStrike: regexp.MustCompile(`^~{2}([\s\S]+?)~{2}`),
	// itemMixed(e.g: ***str***, ~~*str*~~) will be part of the parser
	// or we'll lex recuresively
	itemCode: regexp.MustCompile("^`{1,2}\\s*([\\s\\S]*?[^`])\\s*`{1,2}"),
	itemBr:   regexp.MustCompile(`^ {2,}\n`),
	// Links
	itemLink:     regexp.MustCompile(fmt.Sprintf(`^!?\[(%s)\]\(%s\)`, reLinkText, reLinkHref)),
	itemAutoLink: regexp.MustCompile(`^<([^ >]+(@|:\/)[^ >]+)>`),
	itemGfmLink:  regexp.MustCompile(`^(https?:\/\/[^\s<]+[^<.,:;"')\]\s])`),
	// Image
	// TODO(Ariel): DRY
	itemImage: regexp.MustCompile(fmt.Sprintf(`^!?\[(%s)\]\(%s\)`, reLinkText, reLinkHref)),
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name    string    // the name of the input; used only for error reports
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	pos     Pos       // current position in the input
	start   Pos       // start position of this item
	width   Pos       // width of last rune read from input
	lastPos Pos       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
	eot     Pos       // end of table
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
	case '*', '-', '_', '+':
		p := l.peek()
		if p == '*' || p == '-' || p == '_' {
			l.backup()
			return lexHr
		} else {
			l.backup()
			return lexList
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		l.backup()
		return lexList
	case '>':
		l.emit(itemBlockQuote)
		return lexText
	case '#':
		l.backup()
		return lexHeading
	case ' ', '\t':
		// Should be here ?
		// TODO(Ariel): test that it's a codeBlock and not list for sure
		if block[itemCodeBlock].MatchString(l.input[l.pos-1:]) {
			l.backup()
			return lexCode
		}
		// Keep moving forward until we get all the
		// indentation size
		for ; r == l.peek(); r = l.next() {
		}
		l.emit(itemIndent)
		return lexAny
	case '`', '~':
		// if it's gfm-code
		c := l.input[l.pos : l.pos+2]
		if c == "``" || c == "~~" {
			l.backup()
			return lexGfmCode
		}
		fallthrough
	case '|':
		if m := block[itemLpTable].FindString(l.input[l.pos-1:]); m != "" {
			l.emit(itemLpTable)
			l.eot = l.pos + Pos(len(m))
		}
		fallthrough
	default:
		if m := block[itemTable].FindString(l.input[l.pos-1:]); m != "" {
			l.emit(itemTable)
			l.eot = l.pos + Pos(len(m))
			// we go one step back to get the full text
			// in the lexText phase
			l.start--
		}
		l.backup()
		return lexText
	}
}

// lexHeading scans heading items.
func lexHeading(l *lexer) stateFn {
	if m := block[itemHeading].FindString(l.input[l.pos:]); m != "" {
		// Emit without the newline(\n)
		l.pos += Pos(len(m))
		// TODO(Ariel): hack, fix regexp
		if strings.HasSuffix(m, "\n") {
			l.pos--
		}
		l.emit(itemHeading)
		return lexAny
	}
	return lexText
}

// lexHr scans horizontal rules items.
func lexHr(l *lexer) stateFn {
	if block[itemHr].MatchString(l.input[l.pos:]) {
		match := block[itemHr].FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemHr)
		return lexAny
	}
	return lexText
}

// lexGfmCode scans GFM code block.
func lexGfmCode(l *lexer) stateFn {
	re := block[itemGfmCodeBlock]
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
	match := block[itemCodeBlock].FindString(l.input[l.pos:])
	l.pos += Pos(len(match))
	l.emit(itemCodeBlock)
	return lexAny
}

// lexList scans ordered and unordered lists.
func lexList(l *lexer) stateFn {
	if m := block[itemList].FindString(l.input[l.pos:]); m != "" {
		l.pos += Pos(len(m))
		l.emit(itemList)
	}
	return lexText
}

// lexText scans until eol(\n)
// We have a lot of things to do in this lextext
// for example: ignore itemBr on list/tables
// fix the text scaning etc...
func lexText(l *lexer) stateFn {
	// Drain text before emitting
	emit := func(item itemType, pos Pos) {
		if l.pos > l.start {
			l.emit(itemText)
		}
		l.pos += pos
		l.emit(item)
	}
Loop:
	for {
		switch r := l.peek(); {
		case r == eof:
			emit(eof, Pos(0))
			break Loop
		case r == '\n':
			emit(itemNewLine, l.width)
			break Loop
		case r == ' ':
			if m := span[itemBr].FindString(l.input[l.pos:]); m != "" {
				// pos - length of new-line
				emit(itemBr, Pos(len(m)))
				break
			}
			l.next()
		// if it's start as an emphasis
		case r == '_', r == '*', r == '~', r == '`':
			input := l.input[l.pos:]
			// Strong
			if m := span[itemStrong].FindString(input); m != "" {
				emit(itemStrong, Pos(len(m)))
				break
			}
			// Italic
			if m := span[itemItalic].FindString(input); m != "" {
				emit(itemItalic, Pos(len(m)))
				break
			}
			// Strike
			if m := span[itemStrike].FindString(input); m != "" {
				emit(itemStrike, Pos(len(m)))
				break
			}
			// InlineCode
			if m := span[itemCode].FindString(input); m != "" {
				emit(itemCode, Pos(len(m)))
				break
			}
			l.next()
		// itemLink, itemAutoLink, itemImage
		case r == '[', r == '<', r == '!':
			input := l.input[l.pos:]
			if m := span[itemLink].FindString(input); m != "" {
				pos := Pos(len(m))
				if r == '[' {
					emit(itemLink, pos)
				} else {
					emit(itemImage, pos)
				}
				break
			}
			if m := span[itemAutoLink].FindString(input); m != "" {
				emit(itemAutoLink, Pos(len(m)))
				break
			}
			l.next()
		case r == '|':
			if l.eot > l.pos {
				emit(itemPipe, l.width)
				break
			}
			l.next()
		default:
			input := l.input[l.pos:]
			// Test for Setext-style headers
			if m := block[itemLHeading].FindString(input); m != "" {
				emit(itemLHeading, Pos(len(m)))
				break Loop
			}
			// GfmLink
			if m := span[itemGfmLink].FindString(input); m != "" {
				emit(itemGfmLink, Pos(len(m)))
				break
			}
			l.next()
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

// lexItem return the next item token, clled by the parser.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = l.pos
	return item
}
