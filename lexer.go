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
	itemNewLine
	itemHTML
	itemDefLink
	// Block Elements
	itemHeading
	itemLHeading // Setext-style headers
	itemBlockQuote
	itemList
	itemListItem
	itemLooseItem
	itemCodeBlock
	itemGfmCodeBlock
	itemHr
	itemTable
	itemLpTable
	// Span Elements
	itemText
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
	reEmphasise = `(?s)^_{%[1]d}(.+?(?:_{0,}))_{%[1]d}|^\*{%[1]d}(.+?(?:\*{0,}))\*{%[1]d}`
	reGfmCode   = `(?s)^%[1]s{3,} *(\S+)? *\n(.+?)\s*%[1]s{3,}$*(?:\n+|$)`
	reLinkText  = `(?:\[[^\]]*\]|[^\[\]]|\])*`
	reLinkHref  = `(?s)\s*<?(.*?)>?(?:\s+['"](.*?)['"])?\s*`
	reDefLink   = `^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +["(]([^\n]+)[")])? *(?:\n+|$)`
)

// Block Grammer
var block = map[itemType]*regexp.Regexp{
	itemDefLink:   regexp.MustCompile(reDefLink),
	itemHeading:   regexp.MustCompile(`^ *(#{1,6}) +([^\n]+?) *#* *(?:\n+|$)`),
	itemLHeading:  regexp.MustCompile(`^([^\n]+)\n *(=|-){2,} *(?:\n+|$)`),
	itemHr:        regexp.MustCompile(`^( *[-*_]){3,} *(?:\n+|$)`),
	itemCodeBlock: regexp.MustCompile(`^( {4}[^\n]+\n*)+`),
	// Backreferences is unavailable
	itemGfmCodeBlock: regexp.MustCompile(fmt.Sprintf(reGfmCode, "`") + "|" + fmt.Sprintf(reGfmCode, "~")),
	// itemListRegexp it's actually similar to item, but scan whole line
	itemList:      regexp.MustCompile(`^( *)(?:[*+-]|\d+\.) (.*)(?:\n|)`),
	itemListItem:  regexp.MustCompile(`^ *([*+-]|\d+\.) +`),
	itemLooseItem: regexp.MustCompile(`(?m)\n\n(.*)`),
	// leading-pipe table
	itemLpTable:    regexp.MustCompile(`^ *\|(.+)\n *\|( *[-:]+[-| :]*)\n((?: *\|.*(?:\n|$))*)\n*`),
	itemTable:      regexp.MustCompile(`^ *(\S.*\|.*)\n *([-:]+ *\|[-| :]*)\n((?:.*\|.*(?:\n|$))*)\n*`),
	itemBlockQuote: regexp.MustCompile(`^( *>[^\n]+(\n[^\n]+)*\n*)+`),
}

// Inline Grammer
var span = map[itemType]*regexp.Regexp{
	itemItalic: regexp.MustCompile(fmt.Sprintf(reEmphasise, 1)),
	itemStrong: regexp.MustCompile(fmt.Sprintf(reEmphasise, 2)),
	itemStrike: regexp.MustCompile(`(?s)^~{2}(.+?)~{2}`),
	itemCode:   regexp.MustCompile("(?s)^`{1,2}\\s*(.*?[^`])\\s*`{1,2}"),
	itemBr:     regexp.MustCompile(`^ {2,}\n`),
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
	case eof:
		return nil
	case '*', '-', '_', '+':
		l.next()
		if p := l.peek(); p == '*' || p == '-' || p == '_' {
			l.backup()
			return lexHr
		}
		l.backup()
		return lexList
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return lexList
	case '<':
		return lexHtml
	case '>':
		return lexBlockQuote
	case '[':
		return lexDefLink
	case '#':
		return lexHeading
	case ' ':
		// Should be here ?
		// TODO(Ariel): test that it's a codeBlock and not list for sure
		if block[itemCodeBlock].MatchString(l.input[l.pos:]) {
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
			return lexGfmCode
		}
		fallthrough
	case '|':
		if m := block[itemLpTable].FindString(l.input[l.pos:]); m != "" {
			l.eot = l.start + Pos(len(m))
			l.emit(itemLpTable)
		}
		fallthrough
	default:
		if m := block[itemTable].FindString(l.input[l.pos:]); m != "" {
			l.eot = l.start + Pos(len(m)) - l.width
			l.emit(itemTable)
			// we go one step back to get the full text
			// in the lexText phase
			l.start--
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

// lexText scans until end-of-line(\n)
// We have a lot of things to do in this lextext
// for example: ignore itemBr on list/tables
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
		case r == '|':
			if l.eot > l.pos {
				emit(itemPipe, l.width)
				break
			}
			l.next()
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

// lexItem return the next item token, clled by the parser.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = l.pos
	return item
}

// One phase lexing(inline reason)
func (l *lexer) lexInline() {
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
			// I don;t want to emit EOF(in inline mode)
			// emit(eof, Pos(0))
			l.emit(itemText)
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
		default:
			input := l.input[l.pos:]
			if m := span[itemGfmLink].FindString(input); m != "" {
				emit(itemGfmLink, Pos(len(m)))
				break
			}
			l.next()
		}
	}
	close(l.items)
}

// lexHtml.
func lexHtml(l *lexer) stateFn {
	if match, res := l.MatchHtml(l.input[l.pos:]); match {
		l.pos += Pos(len(res))
		l.emit(itemHTML)
		return lexAny
	}
	return lexText
}

// Test if the given input is match the HTML pattern(blocks only)
func (l *lexer) MatchHtml(input string) (bool, string) {
	comment := regexp.MustCompile(`(?s)<!--.*?-->`)
	if m := comment.FindString(input); m != "" {
		return true, m
	}
	reStart := regexp.MustCompile(`^<(\w+)(?:"[^"]*"|'[^']*'|[^'">])*?>`)
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
	match, items := l.MatchList(l.input[l.pos:])
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
		l.emit(typ, item)
	}
	return lexAny
}

func (l *lexer) MatchList(input string) (bool, []string) {
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
func (l *lexer) MatchBlockQuote(input string) (bool, string) {
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
	if match, res := l.MatchBlockQuote(l.input[l.pos:]); match {
		l.pos += Pos(len(res))
		l.emit(itemBlockQuote)
		return lexAny
	}
	return lexText
}
