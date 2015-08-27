package mark

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parse holds the state of the parser.
type parse struct {
	Nodes     []Node
	lex       Lexer
	options   *Options
	tr        *parse
	output    string
	peekCount int
	token     [3]item                 // three-token lookahead for parser
	links     map[string]*DefLinkNode // Deflink parsing, used RefLinks
	renderFn  map[NodeType]RenderFn   // Custom overridden fns
}

// Return new parser
func newParse(input string, opts *Options) *parse {
	return &parse{
		lex:      lex(input),
		options:  opts,
		links:    make(map[string]*DefLinkNode),
		renderFn: make(map[NodeType]RenderFn),
	}
}

// parse convert the raw text to Nodeparse.
func (p *parse) parse() {
Loop:
	for {
		var n Node
		switch t := p.peek(); t.typ {
		case itemEOF, itemError:
			break Loop
		case itemNewLine:
			p.next()
		case itemHr:
			n = p.newHr(p.next().pos)
		case itemHTML:
			t = p.next()
			n = p.newHTML(t.pos, t.val)
		case itemDefLink:
			n = p.parseDefLink()
		case itemHeading, itemLHeading:
			n = p.parseHeading()
		case itemCodeBlock, itemGfmCodeBlock:
			n = p.parseCodeBlock()
		case itemList:
			n = p.parseList()
		case itemTable, itemLpTable:
			n = p.parseTable()
		case itemBlockQuote:
			n = p.parseBlockQuote()
		case itemIndent:
			space := p.next()
			// If it isn't followed by
			if p.peek().typ != itemText {
				continue
			}
			p.backup2(space)
			fallthrough
		// itemText
		default:
			tmp := p.newParagraph(t.pos)
			tmp.Nodes = p.parseText(p.next().val)
			n = tmp
		}
		if n != nil {
			p.append(n)
		}
	}
}

// Root getter
func (p *parse) root() *parse {
	if p.tr == nil {
		return p
	}
	return p.tr.root()
}

// Render parse nodes to the wanted output
func (p *parse) render() {
	for i, node := range p.Nodes {
		// If there's a custom render function, use it instead.
		if fn, ok := p.renderFn[node.Type()]; ok {
			p.output += fn(node)
		} else {
			p.output += node.Render()
		}
		if i != len(p.Nodes)-1 {
			p.output += "\n"
		}
	}
}

// append new node to nodes-list
func (p *parse) append(n Node) {
	p.Nodes = append(p.Nodes, n)
}

// next returns the next token
func (p *parse) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// peek returns but does not consume the next token.
func (p *parse) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

// backup backs the input stream tp one token
func (p *parse) backup() {
	p.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (p *parse) backup2(t1 item) {
	p.token[1] = t1
	p.peekCount = 2
}

// parseText
func (p *parse) parseText(input string) (nodes []Node) {
	// HACK: if there's more 'itemText' in the way, make it one.
	for {
		tkn := p.next()
		if tkn.typ == itemText {
			input += tkn.val
		} else if tkn.typ == itemNewLine {
			if p.peek().typ != itemText {
				p.backup2(tkn)
				break
			}
			input += tkn.val
		} else {
			p.backup()
			break
		}
	}
	// Removing initial and final whitespace.
	input = regexp.MustCompile(`(?m)^ +| +(\n|$)`).ReplaceAllStringFunc(input, func(s string) string {
		if span[itemBr].MatchString(s) {
			return s
		}
		return ""
	})
	l := lexInline(input)
	for token := range l.items {
		var node Node
		switch token.typ {
		case itemBr:
			node = p.newBr(token.pos)
		case itemStrong, itemItalic, itemStrike, itemCode:
			node = p.parseEmphasis(token.typ, token.pos, token.val)
		case itemLink, itemAutoLink, itemGfmLink:
			var title, text, href string
			match := span[token.typ].FindStringSubmatch(token.val)
			if token.typ == itemLink {
				text, href, title = match[1], match[2], match[3]
			} else {
				text, href = match[1], match[1]
			}
			node = p.newLink(token.pos, title, href, text)
		case itemImage:
			match := span[token.typ].FindStringSubmatch(token.val)
			node = p.newImage(token.pos, match[3], match[2], match[1])
		case itemRefLink, itemRefImage:
			match := span[itemRefLink].FindStringSubmatch(token.val)
			node = p.newRef(token.typ, token.pos, token.val, match[1], match[2])
		// itemText
		default:
			node = p.newText(token.pos, token.val)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// parse inline emphasis
func (p *parse) parseEmphasis(typ itemType, pos Pos, val string) *EmphasisNode {
	node := p.newEmphasis(pos, typ)
	match := span[typ].FindStringSubmatch(val)
	text := match[len(match)-1]
	if text == "" {
		text = match[1]
	}
	node.Nodes = p.parseText(text)
	return node
}

// parse heading block
func (p *parse) parseHeading() (node *HeadingNode) {
	token := p.next()
	match := block[token.typ].FindStringSubmatch(token.val)
	if token.typ == itemHeading {
		node = p.newHeading(token.pos, len(match[1]), match[2])
	} else {
		// using equal signs for first-level, and dashes for second-level.
		level := 1
		if match[2] == "-" {
			level = 2
		}
		node = p.newHeading(token.pos, level, match[1])
	}
	return
}

func (p *parse) parseDefLink() *DefLinkNode {
	token := p.next()
	match := block[itemDefLink].FindStringSubmatch(token.val)
	name := strings.ToLower(match[1])
	// name(lowercase), href, title
	n := p.newDefLink(token.pos, name, match[2], match[3])
	// store in links
	links := p.root().links
	if _, ok := links[name]; !ok {
		links[name] = n
	}
	return n
}

// parse codeBlock
func (p *parse) parseCodeBlock() *CodeNode {
	var lang, text string
	token := p.next()
	if token.typ == itemGfmCodeBlock {
		match := block[itemGfmCodeBlock].FindStringSubmatch(token.val)
		if text = match[2]; text == "" {
			text = match[4]
		}
		if lang = match[1]; lang == "" {
			lang = match[3]
		}
	} else {
		text = regexp.MustCompile("(?m)^( {0,4})").ReplaceAllLiteralString(token.val, "")
	}
	return p.newCode(token.pos, lang, text)
}

func (p *parse) parseBlockQuote() (n *BlockQuoteNode) {
	token := p.next()
	// replacer
	re := regexp.MustCompile(`(?m)^> ?`)
	raw := re.ReplaceAllString(token.val, "")
	// TODO(Ariel): not work right now with defLink(inside the blockQuote)
	tr := &parse{lex: lex(raw), tr: p}
	tr.parse()
	n = p.newBlockQuote(token.pos)
	n.Nodes = tr.Nodes
	return
}

// parse list
func (p *parse) parseList() *ListNode {
	token := p.next()
	list := p.newList(token.pos, isDigit(token.val))
Loop:
	for {
		switch token = p.peek(); token.typ {
		case itemLooseItem, itemListItem:
			list.append(p.parseListItem())
		default:
			break Loop
		}
	}
	return list
}

// parse listItem
// Add ignore list(e.g: table should parse as a text)
func (p *parse) parseListItem() *ListItemNode {
	token := p.next()
	item := p.newListItem(token.pos)
	tr := &parse{lex: lex(strings.TrimSpace(token.val)), tr: p}
	tr.parse()
	for _, node := range tr.Nodes {
		// wrap with paragraph only when it's loose item
		if n, ok := node.(*ParagraphNode); ok && token.typ == itemListItem {
			item.Nodes = append(item.Nodes, n.Nodes...)
		} else {
			item.append(node)
		}
	}
	return item
}

// parse table
func (p *parse) parseTable() *TableNode {
	table := p.newTable(p.next().pos)
	// Align	[ None, Left, Right, ... ]
	// Header	[ Cells: [ ... ] ]
	// Data:	[ Rows: [ Cells: [ ... ] ] ]
	rows := struct {
		Align  []AlignType
		Header []item
		Cells  [][]item
	}{}
Loop:
	for i := 0; ; {
		switch token := p.next(); token.typ {
		case itemTableRow:
			i++
			if i > 2 {
				rows.Cells = append(rows.Cells, []item{})
			}
		case itemTableCell:
			// Header
			if i == 1 {
				rows.Header = append(rows.Header, token)
				// Alignment
			} else if i == 2 {
				rows.Align = append(rows.Align, parseAlign(token.val))
				// Data
			} else {
				pos := i - 3
				rows.Cells[pos] = append(rows.Cells[pos], token)
			}
		default:
			p.backup()
			break Loop
		}
	}
	// Tranform to nodes
	table.append(p.parseCells(Header, rows.Header, rows.Align))
	// Table body
	for _, row := range rows.Cells {
		table.append(p.parseCells(Data, row, rows.Align))
	}
	return table
}

// parse cells and return new row
func (p *parse) parseCells(kind int, items []item, align []AlignType) *RowNode {
	var row *RowNode
	for i, item := range items {
		if i == 0 {
			row = p.newRow(item.pos)
		}
		cell := p.newCell(item.pos, kind, align[i])
		cell.Nodes = p.parseText(item.val)
		row.append(cell)
	}
	return row
}

// get align-string and return the align type of it
// e.g: ":---", "---:", ":---:", "---"
func parseAlign(s string) (typ AlignType) {
	sfx, pfx := strings.HasSuffix(s, ":"), strings.HasPrefix(s, ":")
	switch {
	case sfx && pfx:
		typ = Center
	case sfx:
		typ = Right
	case pfx:
		typ = Left
	}
	return
}

// test if given string is digit
func isDigit(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsDigit(r)
}
