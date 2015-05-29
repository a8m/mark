package mark

import (
	fmt "github.com/k0kubun/pp"
	"regexp"
	"unicode"
	"unicode/utf8"
)

type Tree struct {
	text  string
	lex   *lexer
	Nodes []Node
	// Parsing only
	token     [3]item // three-token lookahead for parser
	peekCount int
	output    string
}

// Parse convert the raw text to NodeTree.
func (t *Tree) parse() {
Loop:
	for {
		var n Node
		switch p := t.peek().typ; p {
		case eof, itemError:
			break Loop
		case itemNewLine:
			n = t.newLine(t.next().pos)
		case itemBr:
			n = t.newBr(t.next().pos)
		case itemHr:
			n = t.newHr(t.next().pos)
		case itemText, itemStrong, itemItalic, itemStrike, itemCode,
			itemLink, itemAutoLink, itemGfmLink, itemImage:
			n = t.parseParagraph()
		case itemHeading, itemLHeading:
			n = t.parseHeading()
		case itemCodeBlock, itemGfmCodeBlock:
			n = t.parseCodeBlock()
		case itemList:
			// 0 for the depth
			n = t.parseList(0)
		default:
			fmt.Println("Nothing to do", p)
		}
		t.append(n)
	}
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

// parseParagraph scan until itemBr occur.
func (t *Tree) parseParagraph() *ParagraphNode {
	token := t.next()
	p := t.newParagraph(token.pos)
Loop:
	for {
		var node Node
		switch token.typ {
		case eof, itemError, itemHeading, itemList, itemIndent:
			t.backup()
			break Loop
		case itemNewLine:
			// Two or more lines continuosly, or block below
			if typ := t.peek().typ; typ == itemNewLine || isBlock(typ) {
				t.backup2(token)
				break Loop
			}
			node = t.newLine(token.pos)
		case itemBr:
			node = t.newBr(token.pos)
		case itemText:
			node = t.newText(token.pos, token.val)
		case itemStrong, itemItalic, itemStrike, itemCode:
			match := span[token.typ].FindStringSubmatch(token.val)
			text := match[len(match)-1]
			if text == "" {
				text = match[1]
			}
			node = t.newEmphasis(token.pos, token.typ, text)
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
		}
		p.append(node)
		token = t.next()
	}
	return p
}

// parse heading block
func (t *Tree) parseHeading() (node *HeadingNode) {
	token := t.next()
	match := block[token.typ].FindStringSubmatch(token.val)
	if token.typ == itemHeading {
		node = t.newHeading(token.pos, len(match[1]), match[2])
	} else {
		// itemLHeading will always be in level 1.
		node = t.newHeading(token.pos, 1, match[1])
	}
	return
}

// parse codeBlock
func (t *Tree) parseCodeBlock() *CodeNode {
	var lang, text string
	token := t.next()
	if token.typ == itemGfmCodeBlock {
		match := block[itemGfmCodeBlock].FindStringSubmatch(token.val)
		lang, text = match[1], match[2]
		if text == "" {
			text = match[4]
		}
	} else {
		text = regexp.MustCompile("(?m)( {4}|\t)").ReplaceAllLiteralString(token.val, "")
	}
	return t.newCode(token.pos, lang, text)
}

// parse list
func (t *Tree) parseList(depth int) *ListNode {
	token := t.next()
	list := t.newList(token.pos, depth, isDigit(token.val))
	item := new(ListItemNode)
Loop:
	for {
		switch token = t.next(); token.typ {
		case eof, itemError:
			break Loop
		// It's actually a listItem
		case itemList:
			// List, but not the same type
			if list.Ordered != isDigit(token.val) || depth > 0 {
				t.backup()
				break Loop
			}
			item = t.parseListItem(token.pos, list)
		case itemNewLine:
			if t.peek().typ == itemNewLine {
				break Loop
			}
			fallthrough
		case itemIndent:
			if depth == len(token.val) {
				item = t.parseListItem(token.pos, list)
			}
			t.backup()
			break Loop
		default:
			t.backup()
			item = t.parseListItem(token.pos, list)
		}
		list.append(item)
	}
	return list
}

func (t *Tree) parseListItem(pos Pos, list *ListNode) *ListItemNode {
	item := t.newListItem(pos, list)
	var n Node
Loop:
	for {
		switch token := t.next(); token.typ {
		case eof, itemError:
			break Loop
		case itemList:
			t.backup()
			break Loop
		case itemNewLine:
			switch typ := t.peek().typ; typ {
			case itemNewLine, eof, itemError, itemList, itemIndent:
				continue
			default:
				n = t.newLine(token.pos)
			}
		case itemIndent:
			if t.peek().typ == itemList {
				depth := len(token.val)
				// If it's in the same depth - sibling
				// or if it's less-than - exit
				if depth <= item.List.Depth {
					t.backup2(token)
					break Loop
				}
				n = t.parseList(depth)
			} else {
				n = t.newText(token.pos, token.val)
			}
		case itemCodeBlock, itemGfmCodeBlock:
			n = t.parseCodeBlock()
		default:
			// DRY
			for _, n := range t.parseParagraph().Nodes {
				if n.Type() != NodeNewLine {
					item.append(n)
				}
			}
			continue
		}
		item.append(n)
	}
	return item
}

// test if given string is digit
func isDigit(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsDigit(r)
}

// test if given token is type block
func isBlock(item itemType) bool {
	switch item {
	case itemHeading, itemLHeading, itemCodeBlock, itemBlockQuote,
		itemList, itemTable, itemGfmCodeBlock, itemHr:
		return true
	default:
		return false
	}
}
