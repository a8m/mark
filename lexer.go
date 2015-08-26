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
	itemEOF
	itemNewLine
	itemHTML
	itemHeading
	itemLHeading
	itemBlockQuote
	itemList
	itemListItem
	itemLooseItem
	itemCodeBlock
	itemGfmCodeBlock
	itemHr
	itemTable
	itemLpTable
	itemTableRow
	itemTableCell
	itemStrong
	itemItalic
	itemStrike
	itemCode
	itemLink
	itemDefLink
	itemRefLink
	itemAutoLink
	itemGfmLink
	itemImage
	itemRefImage
	itemText
	itemBr
	itemPipe
	itemIndent
)

var (
	reEmphasise = `(?s)^_{%[1]d}(.+?(?:_{0,}))_{%[1]d}|^\*{%[1]d}(.+?(?:\*{0,}))\*{%[1]d}`
	reGfmCode   = `(?s)^%[1]s{3,} *(\S+)? *\n(.*?)\s*%[1]s{3,}$*(?:\n+|$)`
	reLinkText  = `(?:\[[^\]]*\]|[^\[\]]|\])*`
	reLinkHref  = `(?s)\s*<?(.*?)>?(?:\s+['"\(](.*?)['"\)])?\s*`
	reDefLink   = `(?s)^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +["'(](.+)['")])? *(?:\n+|$)`
)

// Block Grammer
var block = map[itemType]*regexp.Regexp{
	itemDefLink:      regexp.MustCompile(reDefLink),
	itemHeading:      regexp.MustCompile(`^ *(#{1,6}) +([^\n]+?) *#* *(?:\n+|$)`),
	itemLHeading:     regexp.MustCompile(`^([^\n]+?) *\n *(=|-){1,} *(?:\n+|$)`),
	itemHr:           regexp.MustCompile(`^(?:(?:\* *){3,}|(?:_ *){3,}|(?:- *){3,}) *(?:\n+|$)`),
	itemCodeBlock:    regexp.MustCompile(`^( {4}[^\n]+\n*)+`),
	itemGfmCodeBlock: regexp.MustCompile(fmt.Sprintf(reGfmCode, "`") + "|" + fmt.Sprintf(reGfmCode, "~")),
	itemList:         regexp.MustCompile(`^( *)(?:[*+-]|\d{1,9}\.) (.*)(?:\n|)`),
	itemListItem:     regexp.MustCompile(`^ *([*+-]|\d+\.) +`),
	itemLooseItem:    regexp.MustCompile(`(?m)\n\n(.*)`),
	itemLpTable:      regexp.MustCompile(`(^ *\|.+)\n( *\| *[-:]+[-| :]*)\n((?: *\|.*(?:\n|$))*)\n*`),
	itemTable:        regexp.MustCompile(`^ *(\S.*\|.*)\n *([-:]+ *\|[-| :]*)\n((?:.*\|.*(?:\n|$))*)\n*`),
	itemBlockQuote:   regexp.MustCompile(`^( *>[^\n]*(\n[^\n]+)*\n*)+`),
	itemHTML:         regexp.MustCompile(`^<(\w+)(?:"[^"]*"|'[^']*'|[^'">])*?>`),
}

// Inline Grammer
var span = map[itemType]*regexp.Regexp{
	itemItalic:   regexp.MustCompile(fmt.Sprintf(reEmphasise, 1)),
	itemStrong:   regexp.MustCompile(fmt.Sprintf(reEmphasise, 2)),
	itemStrike:   regexp.MustCompile(`(?s)^~{2}(.+?)~{2}`),
	itemCode:     regexp.MustCompile("(?s)^`{1,2}\\s*(.*?[^`])\\s*`{1,2}"),
	itemBr:       regexp.MustCompile(`^(?: {2,}|\\)\n`),
	itemLink:     regexp.MustCompile(fmt.Sprintf(`^!?\[(%s)\]\(%s\)`, reLinkText, reLinkHref)),
	itemRefLink:  regexp.MustCompile(`^!?\[((?:\[[^\]]*\]|[^\[\]]|\])*)\](?:\s*\[([^\]]*)\])?`),
	itemAutoLink: regexp.MustCompile(`^<([^ >]+(@|:\/)[^ >]+)>`),
	itemGfmLink:  regexp.MustCompile(`^(https?:\/\/[^\s<]+[^<.,:;"')\]\s])`),
	itemImage:    regexp.MustCompile(fmt.Sprintf(`^!?\[(%s)\]\(%s\)`, reLinkText, reLinkHref)),
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// Lexer interface, used to composed it inside the parser
type Lexer interface {
	nextItem() item
}

// lexer holds the state of the scanner.
type lexer struct {
	input   string    // the string being scanned
	state   stateFn   // the next lexing function to enter
	pos     Pos       // current position in the input
	start   Pos       // start position of this item
	width   Pos       // width of last rune read from input
	lastPos Pos       // position of most recent item returned by nextItem
	items   chan item // channel of scanned items
}

// lex creates a new lexer for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// One phase lexing(inline reason)
func lexInline(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.lexInline()
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
	switch r := l.peek(); r {
	case '*', '-', '_':
		return lexHr
	case '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return lexList
	case '<':
		return lexHTML
	case '>':
		return lexBlockQuote
	case '[':
		return lexDefLink
	case '#':
		return lexHeading
	case '`', '~':
		return lexGfmCode
	case ' ':
		// TODO(Ariel): Should be here ?
		if block[itemCodeBlock].MatchString(l.input[l.pos:]) {
			return lexCode
		}
		// Keep moving forward until we get all the indentation size
		for ; r == l.peek(); r = l.next() {
		}
		l.emit(itemIndent)
		return lexAny
	case '|':
		if m := block[itemLpTable].MatchString(l.input[l.pos:]); m {
			l.emit(itemLpTable)
			return lexTable
		}
		fallthrough
	default:
		if m := block[itemTable].MatchString(l.input[l.pos:]); m {
			l.emit(itemTable)
			return lexTable
		}
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
	return lexList
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

// lexText scans until end-of-line(\n)
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
		switch r := l.peek(); r {
		case eof:
			emit(itemEOF, Pos(0))
			break Loop
		case '\n':
			// CM 4.4: An indented code block cannot interrupt a paragraph.
			if l.pos > l.start && strings.HasPrefix(l.input[l.pos+1:], "    ") {
				l.next()
				continue
			}
			emit(itemNewLine, l.width)
			break Loop
		default:
			// Test for Setext-style headers
			if m := block[itemLHeading].FindString(l.input[l.pos:]); m != "" {
				emit(itemLHeading, Pos(len(m)))
				break Loop
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
func (l *lexer) emit(t itemType, s ...string) {
	if len(s) == 0 {
		s = append(s, l.input[l.start:l.pos])
	}
	l.items <- item{t, l.start, s[0]}
	l.start = l.pos
}

// lexItem return the next item token, called by the parser.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = l.pos
	return item
}

// One phase lexing(inline reason)
func (l *lexer) lexInline() {
	escape := regexp.MustCompile("^\\\\([\\`*{}\\[\\]()#+\\-.!_>~|])")
	// Drain text before emitting
	emit := func(item itemType, pos int) {
		if l.pos > l.start {
			l.emit(itemText)
		}
		l.pos += Pos(pos)
		l.emit(item)
	}
Loop:
	for {
		switch r := l.peek(); r {
		case eof:
			if l.pos > l.start {
				l.emit(itemText)
			}
			break Loop
			// backslash escaping
		case '\\':
			if m := escape.FindStringSubmatch(l.input[l.pos:]); len(m) != 0 {
				if l.pos > l.start {
					l.emit(itemText)
				}
				l.pos += Pos(len(m[0]))
				l.emit(itemText, m[1])
				break
			}
			fallthrough
		case ' ':
			if m := span[itemBr].FindString(l.input[l.pos:]); m != "" {
				// pos - length of new-line
				emit(itemBr, len(m))
				break
			}
			l.next()
		// if it's start as an emphasis
		case '_', '*', '~', '`':
			input := l.input[l.pos:]
			// Strong
			if m := span[itemStrong].FindString(input); m != "" {
				emit(itemStrong, len(m))
				break
			}
			// Italic
			if m := span[itemItalic].FindString(input); m != "" {
				emit(itemItalic, len(m))
				break
			}
			// Strike
			if m := span[itemStrike].FindString(input); m != "" {
				emit(itemStrike, len(m))
				break
			}
			// InlineCode
			if m := span[itemCode].FindString(input); m != "" {
				emit(itemCode, len(m))
				break
			}
			l.next()
		// itemLink, itemImage, itemRefLink, itemRefImage
		case '[', '!':
			input := l.input[l.pos:]
			if m := span[itemLink].FindString(input); m != "" {
				pos := len(m)
				if r == '[' {
					emit(itemLink, pos)
				} else {
					emit(itemImage, pos)
				}
				break
			}
			if m := span[itemRefLink].FindString(input); m != "" {
				pos := len(m)
				if r == '[' {
					emit(itemRefLink, pos)
				} else {
					emit(itemRefImage, pos)
				}
				break
			}
			l.next()
		// itemAutoLink,
		case '<':
			if m := span[itemAutoLink].FindString(l.input[l.pos:]); m != "" {
				emit(itemAutoLink, len(m))
				break
			}
			l.next()
		default:
			if m := span[itemGfmLink].FindString(l.input[l.pos:]); m != "" {
				emit(itemGfmLink, len(m))
				break
			}
			l.next()
		}
	}
	close(l.items)
}

// lexHTML.
func lexHTML(l *lexer) stateFn {
	if match, res := l.matchHTML(l.input[l.pos:]); match {
		l.pos += Pos(len(res))
		l.emit(itemHTML)
		return lexAny
	}
	return lexText
}

// Test if the given input is match the HTML pattern(blocks only)
func (l *lexer) matchHTML(input string) (bool, string) {
	// TODO: DRY regexp - multiline comment
	comment := regexp.MustCompile(`(?s)<!--.*?-->`)
	if m := comment.FindString(input); m != "" {
		return true, m
	}
	reStart := block[itemHTML]
	// TODO: Add all span-tags and move to config.
	reSpan := regexp.MustCompile(`^(a|em|strong|small|s|q|data|time|code|sub|sup|i|b|u|span|br|del|img)$`)
	if m := reStart.FindStringSubmatch(input); len(m) != 0 {
		el, name := m[0], m[1]
		// if name is a span... is a text
		if reSpan.MatchString(name) {
			return false, ""
		}
		// if it's a self-closed html element, but not a itemAutoLink
		if strings.HasSuffix(el, "/>") && !span[itemAutoLink].MatchString(el) {
			return true, el
		}
		reStr := fmt.Sprintf(`(?s)(.)+?<\/%s> *(?:\n{2,}|\s*$)`, name)
		reMatch, err := regexp.Compile(reStr)
		if err != nil {
			return false, ""
		}
		if m := reMatch.FindString(input); m != "" {
			return true, m
		}
	}
	return false, ""
}

// lexDefLink scans link definition
func lexDefLink(l *lexer) stateFn {
	if m := block[itemDefLink].FindString(l.input[l.pos:]); m != "" {
		l.pos += Pos(len(m))
		l.emit(itemDefLink)
		return lexAny
	}
	return lexText
}

// lexList scans ordered and unordered lists.
func lexList(l *lexer) stateFn {
	match, items := l.matchList(l.input[l.pos:])
	if !match {
		return lexText
	}
	var space int
	var typ itemType
	reItem := block[itemListItem]
	reLoose := block[itemLooseItem]
	for i, item := range items {
		// Emit itemList on the first loop
		if i == 0 {
			l.emit(itemList, reItem.FindStringSubmatch(item)[1])
		}
		// Initialize each loop
		typ = itemListItem
		space = len(item)
		l.pos += Pos(space)
		item = reItem.ReplaceAllString(item, "")
		// Indented
		if strings.Contains(item, "\n ") {
			space -= len(item)
			reSpace := regexp.MustCompile(fmt.Sprintf(`(?m)^ {1,%d}`, space))
			item = reSpace.ReplaceAllString(item, "")
		}
		// If current is loose
		for _, l := range reLoose.FindAllString(item, -1) {
			if len(strings.TrimSpace(l)) > 0 || i != len(items)-1 {
				typ = itemLooseItem
				break
			}
		}
		// or previous
		if typ != itemLooseItem && i > 0 && strings.HasSuffix(items[i-1], "\n\n") {
			typ = itemLooseItem
		}
		l.emit(typ, strings.TrimSpace(item))
	}
	return lexAny
}

func (l *lexer) matchList(input string) (bool, []string) {
	var res []string
	reItem := block[itemList]
	reScan := regexp.MustCompile(`^(.*)(?:\n|)`)
	reLine := regexp.MustCompile(`^\n{1,}`)
	if !reItem.MatchString(input) {
		return false, res
	}
	// First item
	m := reItem.FindStringSubmatch(input)
	item, depth := m[0], len(m[1])
	input = input[len(item):]
	// Loop over the input
	for len(input) > 0 {
		// Count new-lines('\n')
		if m := reLine.FindString(input); m != "" {
			item += m
			input = input[len(m):]
			if len(m) >= 2 || !reItem.MatchString(input) && !strings.HasPrefix(input, " ") {
				break
			}
		}
		// DefLink or hr
		if block[itemDefLink].MatchString(input) || block[itemHr].MatchString(input) {
			break
		}
		// It's list in the same depth
		if m := reItem.FindStringSubmatch(input); len(m) > 0 && len(m[1]) == depth {
			if item != "" {
				res = append(res, item)
			}
			item = m[0]
			input = input[len(item):]
		} else {
			m := reScan.FindString(input)
			item += m
			input = input[len(m):]
		}
	}
	// Drain res
	if item != "" {
		res = append(res, item)
	}
	return true, res
}

// Test if the given input match blockquote
func (l *lexer) matchBlockQuote(input string) (bool, string) {
	match := block[itemBlockQuote].FindString(input)
	if match == "" {
		return false, match
	}
	lines := strings.Split(match, "\n")
	for i, line := range lines {
		// if line is a link-definition we cut the match until this point
		if isDef := block[itemDefLink].MatchString(line); isDef {
			match = strings.Join(lines[0:i], "\n")
			break
		}
	}
	return true, match
}

// lexBlockQuote
func lexBlockQuote(l *lexer) stateFn {
	if match, res := l.matchBlockQuote(l.input[l.pos:]); match {
		l.pos += Pos(len(res))
		l.emit(itemBlockQuote)
		return lexAny
	}
	return lexText
}

// lexTable
func lexTable(l *lexer) stateFn {
	re := block[itemTable]
	if l.peek() == '|' {
		re = block[itemLpTable]
	}
	table := re.FindStringSubmatch(l.input[l.pos:])
	l.pos += Pos(len(table[0]))
	l.start = l.pos
	// Ignore the first match, and flat all rows(by splitting \n)
	rows := append(table[1:3], strings.Split(table[3], "\n")...)
	trim := regexp.MustCompile(`^ *\| *| *\| *$`)
	split := regexp.MustCompile(` *\| *`)
	// Loop over the rows
	for _, row := range rows {
		if row == "" {
			continue
		}
		l.emit(itemTableRow)
		rawCells := trim.ReplaceAllString(row, "")
		cells := split.Split(rawCells, -1)
		// Emit cells in the current row
		for _, cell := range cells {
			l.emit(itemTableCell, cell)
		}
	}
	return lexAny
}
