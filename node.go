package mark

import (
	"fmt"
	"strconv"
)

// A Node is an element in the parse tree.
type Node interface {
	Type() NodeType
	Render() string
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText NodeType = iota // Plain text.
	NodeParagraph
	NodeEmphasis
	NodeHeading
	NodeNewLine
	NodeHr
	NodeList
	NodeCode       // Code block.
	NodeBlockQuote // Blockquote block.
)

// ParagraphNode hold simple paragraph node contains text
// that may be emphasis.
type ParagraphNode struct {
	NodeType
	Pos
	Nodes []Node
}

// Render return the html representation of ParagraphNode
func (n *ParagraphNode) Render() (s string) {
	for _, node := range n.Nodes {
		s += node.Render()
	}
	return render("p", s)
}

func (t *ParagraphNode) append(n Node) {
	t.Nodes = append(t.Nodes, n)
}

func (t *Tree) newParagraph(pos Pos) *ParagraphNode {
	return &ParagraphNode{NodeType: NodeParagraph, Pos: pos}
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	Text []byte
}

// Render return the string representation of TexNode
func (n *TextNode) Render() string {
	return string(n.Text)
}

func (t *Tree) newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

// NewLineNode represent simple `\n`.
type NewLineNode struct {
	NodeType
	Pos
}

// Render return the string \n for representing new line.
func (n *NewLineNode) Render() string {
	return "\n"
}

func (t *Tree) newLine(pos Pos) *NewLineNode {
	return &NewLineNode{NodeType: NodeNewLine, Pos: pos}
}

// HrNode represent
type HrNode struct {
	NodeType
	Pos
}

// Render return the html representation of hr.
func (n *HrNode) Render() string {
	return "<hr>"
}

func (t *Tree) newHr(pos Pos) *HrNode {
	return &HrNode{NodeType: NodeHr, Pos: pos}
}

// EmphasisNode holds text with style.
type EmphasisNode struct {
	NodeType
	Pos
	Style itemType
	Text  []byte
}

// Tag return the tagName based on Style field
func (n *EmphasisNode) Tag() (s string) {
	switch n.Style {
	case itemStrong:
		s = "strong"
	case itemItalic:
		s = "em"
	case itemStrike:
		s = "del"
	case itemCode:
		s = "code"
	}
	return
}

// Return the html representation of emphasis text(string, italic, ..).
func (n *EmphasisNode) Render() string {
	return render(n.Tag(), string(n.Text))
}

func (t *Tree) newEmphasis(pos Pos, style itemType, text string) *EmphasisNode {
	return &EmphasisNode{NodeType: NodeEmphasis, Pos: pos, Style: style, Text: []byte(text)}
}

// Heading holds heaing node with specific level.
type HeadingNode struct {
	NodeType
	Pos
	Level int
	Text  []byte
}

// Render return the html representation based on heading level.
func (n *HeadingNode) Render() string {
	return render("h"+strconv.Itoa(n.Level), string(n.Text))
}

func (t *Tree) newHeading(pos Pos, level int, text string) *HeadingNode {
	return &HeadingNode{NodeType: NodeHeading, Pos: pos, Level: level, Text: []byte(text)}
}

// Code holds CodeBlock node with specific lang
type CodeNode struct {
	NodeType
	Pos
	Lang string
	Text []byte
}

// Return the html representation of codeBlock
func (n *CodeNode) Render() string {
	attr := fmt.Sprintf(" class=\"lang-%s\"", n.Lang)
	if n.Lang == "" {
		attr = ""
	}
	code := fmt.Sprintf("<%[1]s%s>%s</%[1]s>", "code", attr, n.Text)
	return render("pre", code)
}

func (t *Tree) newCode(pos Pos, lang, text string) *CodeNode {
	return &CodeNode{NodeType: NodeCode, Pos: pos, Lang: lang, Text: []byte(text)}
}

// Wrap text with specific tag.
func render(tag, body string) string {
	return fmt.Sprintf("<%[1]s>%s</%[1]s>", tag, body)
}
