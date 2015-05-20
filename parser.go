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
		case eof, itemError:
			break Loop
		case itemBr:
			t.append(t.newLine(t.next().pos))
		case itemText, itemStrong, itemItalic, itemStrike, itemCode:
			t.parseParagraph()
		default:
			fmt.Println("Nothing to do", p)
		}
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

// parseParagraph scan until itemBr occur.
func (t *Tree) parseParagraph() {
	token := t.next()
	p := t.newParagraph(token.pos)
Loop:
	for {
		var node Node
		switch token.typ {
		case eof, itemError, itemBr:
			t.backup()
			break Loop
		case itemNewLine:
			node = t.newLine(token.pos)
		case itemText:
			node = t.newText(token.pos, token.val)
		case itemStrong, itemItalic, itemStrike, itemCode:
			// TODO(Ariel): Make sure that it works well with all types
			match := span[token.typ].FindStringSubmatch(token.val)
			node = t.newEmphasis(token.pos, token.typ, match[2])
		}
		p.append(node)
		token = t.next()
	}
	t.append(p)
}
