package mark

import (
	fmt "github.com/k0kubun/pp"
	"regexp"
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
		switch p := t.peek().typ; p {
		case eof, itemError:
			break Loop
		case itemBr, itemNewLine:
			t.append(t.newLine(t.next().pos))
		case itemHr:
			t.append(t.newHr(t.next().pos))
		case itemText, itemStrong, itemItalic, itemStrike, itemCode:
			t.parseParagraph()
		case itemHeading:
			t.parseHeading()
		case itemCodeBlock, itemGfmCodeBlock:
			t.parseCodeBlock()
		default:
			fmt.Println("Nothing to do", p)
		}
	}
}

// Render parse nodes to the wanted output
func (t *Tree) render() {
	// wrap with html/xhtml/head(with options) etc..
	for _, node := range t.Nodes {
		t.output += node.Render()
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
			match := span[token.typ].FindStringSubmatch(token.val)
			node = t.newEmphasis(token.pos, token.typ, match[len(match)-1])
		}
		p.append(node)
		token = t.next()
	}
	t.append(p)
}

// parse heading block
// TODO(Ariel): itemLHeading
func (t *Tree) parseHeading() {
	token := t.next()
	match := block[token.typ].FindStringSubmatch(token.val)
	t.append(t.newHeading(token.pos, len(match[1]), match[2]))
}

// parse codeBlock
func (t *Tree) parseCodeBlock() {
	var lang, text string
	token := t.next()
	if token.typ == itemGfmCodeBlock {
		match := block[itemGfmCodeBlock].FindStringSubmatch(token.val)
		lang, text = match[1], match[2]
	} else {
		text = regexp.MustCompile("(?m) {4}").ReplaceAllLiteralString(token.val, "")
	}
	t.append(t.newCode(token.pos, lang, text))
}
