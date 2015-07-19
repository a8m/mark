package mark

import (
	fmt "github.com/k0kubun/pp"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Tree struct {
	lex   *lexer
	tr    *Tree
	Nodes []Node
	// Parsing only
	token     [3]item // three-token lookahead for parser
	peekCount int
	output    string
	links     map[string]*DefLinkNode
}

// Parse convert the raw text to NodeTree.
func (t *Tree) parse() {
Loop:
	for {
		var n Node
		switch p := t.peek(); p.typ {
		case itemEOF, itemError:
			break Loop
		case itemNewLine:
			n = t.newLine(t.next().pos)
		case itemHr:
			n = t.newHr(t.next().pos)
		case itemHTML:
			p = t.next()
			n = t.newHTML(p.pos, p.val)
		case itemDefLink:
			n = t.parseDefLink()
		case itemHeading, itemLHeading:
			n = t.parseHeading()
		case itemCodeBlock, itemGfmCodeBlock:
			n = t.parseCodeBlock()
		case itemList:
			n = t.parseList()
		case itemTable, itemLpTable:
			n = t.parseTable()
		case itemBlockQuote:
			n = t.parseBlockQuote()
		case itemIndent:
			space := t.next()
			// If it's no follow by text
			if t.peek().typ != itemText {
				continue
			}
			t.backup2(space)
			fallthrough
		case itemText:
			tmp := t.newParagraph(p.pos)
			tmp.Nodes = t.parseText(t.next().val)
			n = tmp
		default:
			t.next()
		}
		if n != nil {
			t.append(n)
		}
	}
}

// Root getter
func (t *Tree) root() *Tree {
	if t.tr == nil {
		return t
	}
	return t.tr.root()
}

// Render parse nodes to the wanted output
func (t *Tree) render() {
	var last Node
	last = t.newLine(0)
	for _, node := range t.Nodes {
		if last.Type() != NodeNewLine || node.Type() != last.Type() {
			t.output += node.Render()
		}
		last = node
	}
}

// append new node to nodes-list
func (t *Tree) append(n Node) {
	t.Nodes = append(t.Nodes, n)
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

// backup backs the input stream tp one token
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// parseText scan until itemBr occur.
func (t *Tree) parseText(input string) (nodes []Node) {
	// HACK: if there's more 'itemText' in the way, make it one.
	for {
		tkn := t.next()
		if tkn.typ == itemText {
			input += tkn.val
		} else if tkn.typ == itemNewLine {
			if t.peek().typ != itemText {
				t.backup2(tkn)
				break
			}
			input += tkn.val
		} else {
			t.backup()
			break
		}
	}
	l := lexInline(input)
	for token := range l.items {
		var node Node
		switch token.typ {
		case itemBr:
			node = t.newBr(token.pos)
		case itemStrong, itemItalic, itemStrike, itemCode:
			node = t.parseEmphasis(token.typ, token.pos, token.val)
		case itemLink, itemAutoLink, itemGfmLink:
			var title, text, href string
			match := span[token.typ].FindStringSubmatch(token.val)
			if token.typ == itemLink {
				text, href, title = match[1], match[2], match[3]
			} else {
				text, href = match[1], match[1]
			}
			node = t.newLink(token.pos, title, href, text)
		case itemImage:
			match := span[token.typ].FindStringSubmatch(token.val)
			node = t.newImage(token.pos, match[3], match[2], match[1])
		case itemRefLink, itemRefImage:
			match := span[itemRefLink].FindStringSubmatch(token.val)
			node = t.newRef(token.typ, token.pos, token.val, match[1], match[2])
		case itemText:
			node = t.newText(token.pos, token.val)
		default:
			fmt.Println("Matching not found for this token:", token)
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// Parse inline emphasis
func (t *Tree) parseEmphasis(typ itemType, pos Pos, val string) *EmphasisNode {
	node := t.newEmphasis(pos, typ)
	match := span[typ].FindStringSubmatch(val)
	text := match[len(match)-1]
	if text == "" {
		text = match[1]
	}
	node.Nodes = t.parseText(text)
	return node
}

// parse heading block
func (t *Tree) parseHeading() (node *HeadingNode) {
	token := t.next()
	match := block[token.typ].FindStringSubmatch(token.val)
	if token.typ == itemHeading {
		node = t.newHeading(token.pos, len(match[1]), match[2])
	} else {
		// using equal signs for first-level, and dashes for second-level.
		level := 1
		if match[2] == "-" {
			level = 2
		}
		node = t.newHeading(token.pos, level, match[1])
	}
	return
}

func (t *Tree) parseDefLink() *DefLinkNode {
	token := t.next()
	match := block[itemDefLink].FindStringSubmatch(token.val)
	name := strings.ToLower(match[1])
	// name(lowercase), href, title
	n := t.newDefLink(token.pos, name, match[2], match[3])
	// store in links
	t.links[name] = n
	return n
}

// parse codeBlock
func (t *Tree) parseCodeBlock() *CodeNode {
	var lang, text string
	token := t.next()
	if token.typ == itemGfmCodeBlock {
		match := block[itemGfmCodeBlock].FindStringSubmatch(token.val)
		if text = match[2]; text == "" {
			text = match[4]
		}
		if lang = match[1]; lang == "" {
			lang = match[3]
		}
	} else {
		text = regexp.MustCompile("(?m)^( {4})").ReplaceAllLiteralString(token.val, "")
	}
	return t.newCode(token.pos, lang, text)
}

func (t *Tree) parseBlockQuote() (n *BlockQuoteNode) {
	token := t.next()
	// replacer
	re := regexp.MustCompile(`(?m)^> ?`)
	raw := re.ReplaceAllString(token.val, "")
	// TODO(Ariel): not work right now with defLink(inside the blockQuote)
	tr := &Tree{lex: lex(raw)}
	tr.parse()
	n = t.newBlockQuote(token.pos)
	n.Nodes = tr.Nodes
	return
}

// parse list
func (t *Tree) parseList() *ListNode {
	token := t.next()
	list := t.newList(token.pos, isDigit(token.val))
Loop:
	for {
		switch token = t.peek(); token.typ {
		case itemLooseItem, itemListItem:
			list.append(t.parseListItem())
		default:
			break Loop
		}
	}
	return list
}

// parse listItem
// Add ignore list(e.g: table should parse as a text)
func (t *Tree) parseListItem() *ListItemNode {
	token := t.next()
	item := t.newListItem(token.pos)
	tr := &Tree{lex: lex(strings.TrimSpace(token.val))}
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
func (t *Tree) parseTable() *TableNode {
	// End of table
	done := t.lex.eot
	token := t.next()
	// ignore the first and last one...
	//lp := token.val == "|"
	table := t.newTable(token.pos)
	// Align	[ None, Left, Right, ... ]
	// Header	[ Cels: [token, token, ... ] ]
	// Data:	[ Row: [Cells: [token, ... ] ] ]
	rows := struct {
		Align  []AlignType
		Header [][]item
		Data   [][][]item
	}{}
	var cell []item
	var row [][]item
	// Collect items
Loop:
	for i := 0; ; {
		switch token := t.next(); token.typ {
		case itemEOF, itemError:
			break Loop
		case itemNewLine:
			// If we done with this table
			if t.peek().pos >= done {
				break Loop
			}
			fallthrough
		case itemPipe:
			// Test if cell non-empty before appending to current row
			if len(cell) > 0 {
				// Header
				if i == 0 {
					rows.Header = append(rows.Header, cell)
					// Alignment
				} else if i == 1 {
					align := cell[0].val
					if len(cell) > 1 {
						for i := 1; i < len(cell); i++ {
							align += cell[i].val
						}
					}
					// Trim spaces
					rows.Align = append(rows.Align, parseAlign(align))
					// Data
				} else {
					row = append(row, cell)
				}
			}
			if token.typ == itemNewLine {
				i++
				// test if there's an elemnts to append to tbody.
				// we want to avoid situations like `appending empty rows`, etc..
				if i > 2 && len(row) > 0 {
					rows.Data = append(rows.Data, row)
					row = [][]item{}
				}
			}
			cell = []item{}
		default:
			cell = append(cell, token)
		}
	}
	// Drain cell/row
	if len(cell) > 0 {
		row = append(row, cell)
	}
	if len(row) > 0 {
		rows.Data = append(rows.Data, row)
	}
	// Tranform to nodes
	// Add an average mechanisem, that ignore empty(or " ") in end of cell
	//	rowLen := len(rows.Align)
	// Table head
	table.append(t.parseCells(Header, rows.Header, rows.Align))
	// Table body
	for _, row := range rows.Data {
		table.append(t.parseCells(Data, row, rows.Align))
	}
	return table
}

// Should return typ []CellNode
func (t *Tree) parseCells(kind int, items [][]item, align []AlignType) *RowNode {
	// TODO(Ariel): Add real position
	row := t.newRow(1)
	for i, item := range items {
		// Cell contain nodes
		var s string
		cell := t.newCell(item[0].pos, kind, align[i])
		for _, tkn := range item {
			s += tkn.val
		}
		// Used before: `^[ |\t]*|[ |\t]*$`
		cell.Nodes = t.parseText(strings.TrimSpace(s))
		row.append(cell)
	}
	return row
}

// get align-string and return the align type of it
// e.g: ":---", "---:", ":---:", "---"
func parseAlign(s string) (typ AlignType) {
	// Trim spaces before
	s = strings.TrimSpace(s)
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
