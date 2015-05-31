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
	NodeBr
	NodeHr
	NodeImage
	NodeList
	NodeListItem
	NodeCode // Code block.
	NodeLink
	NodeTable
	NodeRow
	NodeCell
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

// HrNode represent horizontal rule
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

// BrNode represent br element
type BrNode struct {
	NodeType
	Pos
}

// Render return the html representation of br.
func (n *BrNode) Render() string {
	return "<br>"
}

func (t *Tree) newBr(pos Pos) *BrNode {
	return &BrNode{NodeType: NodeBr, Pos: pos}
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
	var attr string
	if n.Lang != "" {
		attr = fmt.Sprintf(" class=\"lang-%s\"", n.Lang)
	}
	code := fmt.Sprintf("<%[1]s%s>%s</%[1]s>", "code", attr, n.Text)
	return render("pre", code)
}

func (t *Tree) newCode(pos Pos, lang, text string) *CodeNode {
	return &CodeNode{NodeType: NodeCode, Pos: pos, Lang: lang, Text: []byte(text)}
}

// Link holds a tag with optional title
type LinkNode struct {
	NodeType
	Pos
	Title string
	Href  string
	Text  []byte
}

// Return the html representation of link node
func (n *LinkNode) Render() string {
	attrs := fmt.Sprintf("href=\"%s\"", n.Href)
	if n.Title != "" {
		attrs += fmt.Sprintf(" title=\"%s\"", n.Title)
	}
	return fmt.Sprintf("<a %s>%s</a>", attrs, n.Text)
}

func (t *Tree) newLink(pos Pos, title, href, text string) *LinkNode {
	return &LinkNode{NodeType: NodeLink, Title: title, Href: href, Text: []byte(text)}
}

// Image holds img tag with optional title
type ImageNode struct {
	NodeType
	Pos
	Title string
	Src   string
	Alt   []byte
}

// Return the html representation on img node
func (n *ImageNode) Render() string {
	attrs := fmt.Sprintf("src=\"%s\" alt=\"%s\"", n.Src, n.Alt)
	if n.Title != "" {
		attrs += fmt.Sprintf(" title=\"%s\"", n.Title)
	}
	return fmt.Sprintf("<img %s>", attrs)
}

func (t *Tree) newImage(pos Pos, title, src, alt string) *ImageNode {
	return &ImageNode{NodeType: NodeImage, Pos: pos, Title: title, Src: src, Alt: []byte(alt)}
}

// List holds list items nodes in ordered or unordered states.
type ListNode struct {
	NodeType
	Pos
	Ordered bool
	Depth   int
	Items   []*ListItemNode
}

func (t *ListNode) append(item *ListItemNode) {
	t.Items = append(t.Items, item)
}

// Return the html representation of list(ul|ol)
func (n *ListNode) Render() (s string) {
	tag := "ul"
	if n.Ordered {
		tag = "ol"
	}
	for _, item := range n.Items {
		s += item.Render()
	}
	return render(tag, s)
}

func (t *Tree) newList(pos Pos, depth int, ordered bool) *ListNode {
	return &ListNode{NodeType: NodeList, Pos: pos, Ordered: ordered, Depth: depth}
}

// ListItem represent single item in ListNode that may contains nested nodes.
type ListItemNode struct {
	NodeType
	Pos
	Nodes []Node
	List  *ListNode
}

func (t *ListItemNode) append(n Node) {
	t.Nodes = append(t.Nodes, n)
}

// Return the html representation of listItem
func (n *ListItemNode) Render() (s string) {
	for _, node := range n.Nodes {
		s += node.Render()
	}
	return render("li", s)
}

func (t *Tree) newListItem(pos Pos, list *ListNode) *ListItemNode {
	return &ListItemNode{NodeType: NodeListItem, Pos: pos, List: list}
}

// TableNode represent table elment contains head and body
type TableNode struct {
	NodeType
	Pos
	Rows []*RowNode
}

func (n *TableNode) Render() string {
	return ""
}

func (t *Tree) newTable(pos Pos) *TableNode {
	return &TableNode{NodeType: NodeTable, Pos: pos}
}

// TableRowNode represnt tr that holds batch of table-data/cells
type RowNode struct {
	NodeType
	Pos
	Cells []*CellNode
}

func (n *RowNode) Render() string {
	return ""
}

func (t *Tree) newRow(pos Pos) *RowNode {
	return &RowNode{NodeType: NodeRow, Pos: pos}
}

// Alignment
const (
	None = iota
	Right
	Left
	Center
)

// Cell types
const (
	Header = iota
	Data
)

// TableCellNode represent table-data/cell that holds simple text(my be emphasis)
// Note: the text in <th> elements are bold and centered by default.
type CellNode struct {
	NodeType
	Pos
	Kind, Align int
	Nodes       []Node
}

func (n *CellNode) Render() string {
	return ""
}

func (t *Tree) newCell(pos Pos, kind, align int) *CellNode {
	return &CellNode{NodeType: NodeCell, Pos: pos, Kind: kind, Align: align}
}

// TODO(Ariel): rename to wrap()
// Wrap text with specific tag.
func render(tag, body string) string {
	return fmt.Sprintf("<%[1]s>%s</%[1]s>", tag, body)
}
