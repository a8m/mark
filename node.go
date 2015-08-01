package mark

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// Render function, used for overriding default rendering.
type RenderFn func(Node) string

const (
	NodeText       NodeType = iota // A plain text
	NodeParagraph                  // A Paragraph
	NodeEmphasis                   // An emphasis(strong, em, ...)
	NodeHeading                    // A heading (h1, h2, ...)
	NodeBr                         // A link break
	NodeHr                         // A horizontal rule
	NodeImage                      // An image
	NodeRefImage                   // A image reference
	NodeList                       // A list of ListItems
	NodeListItem                   // A list item node
	NodeLink                       // A link(href)
	NodeRefLink                    // A link reference
	NodeDefLink                    // A link definition
	NodeTable                      // A table of NodeRows
	NodeRow                        // A row of NodeCells
	NodeCell                       // A table-cell(td)
	NodeCode                       // A code block(wrapped with pre)
	NodeBlockQuote                 // A blockquote
	NodeHTML                       // An inline HTML
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
	return wrap("p", s)
}

func (t *ParagraphNode) append(n Node) {
	t.Nodes = append(t.Nodes, n)
}

func (t *Parse) newParagraph(pos Pos) *ParagraphNode {
	return &ParagraphNode{NodeType: NodeParagraph, Pos: pos}
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	Pos
	Text string
}

// Render return the string representation of TexNode
func (n *TextNode) Render() string {
	return escape(n.Text)
}

func (t *Parse) newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: text}
}

// HTMLNode holds the raw html source.
type HTMLNode struct {
	NodeType
	Pos
	Src string
}

// Render return the src of the HTMLNode
func (n *HTMLNode) Render() string {
	return n.Src
}

func (t *Parse) newHTML(pos Pos, src string) *HTMLNode {
	return &HTMLNode{NodeType: NodeHTML, Pos: pos, Src: src}
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

func (t *Parse) newHr(pos Pos) *HrNode {
	return &HrNode{NodeType: NodeHr, Pos: pos}
}

// BrNode represent a link-break element.
type BrNode struct {
	NodeType
	Pos
}

// Render return the html representation of line-break.
func (n *BrNode) Render() string {
	return "<br>"
}

func (t *Parse) newBr(pos Pos) *BrNode {
	return &BrNode{NodeType: NodeBr, Pos: pos}
}

// EmphasisNode holds plain-text wrapped with style.
// (strong, em, del, code)
type EmphasisNode struct {
	NodeType
	Pos
	Style itemType
	Nodes []Node
}

// Tag return the tagName based on the Style field.
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

// Return the html representation of emphasis text.
func (n *EmphasisNode) Render() string {
	var s string
	for _, node := range n.Nodes {
		s += node.Render()
	}
	return wrap(n.Tag(), s)
}

func (n *EmphasisNode) append(node Node) {
	n.Nodes = append(n.Nodes, node)
}

func (t *Parse) newEmphasis(pos Pos, style itemType) *EmphasisNode {
	return &EmphasisNode{NodeType: NodeEmphasis, Pos: pos, Style: style}
}

// Heading holds heaing element with specific level(1-6).
type HeadingNode struct {
	NodeType
	Pos
	Level int
	Text  string
}

// Render return the html representation based on heading level.
func (n *HeadingNode) Render() string {
	text := escape(n.Text)
	re := regexp.MustCompile(`[^\w]+`)
	id := re.ReplaceAllString(text, "-")
	// ToLowerCase
	id = strings.ToLower(id)
	return fmt.Sprintf("<%[1]s id=\"%s\">%s</%[1]s>", "h"+strconv.Itoa(n.Level), id, text)
}

func (t *Parse) newHeading(pos Pos, level int, text string) *HeadingNode {
	return &HeadingNode{NodeType: NodeHeading, Pos: pos, Level: level, Text: text}
}

// Code holds CodeBlock node with specific lang field.
type CodeNode struct {
	NodeType
	Pos
	Lang, Text string
}

// Return the html representation of codeBlock
func (n *CodeNode) Render() string {
	var attr string
	if n.Lang != "" {
		attr = fmt.Sprintf(" class=\"lang-%s\"", n.Lang)
	}
	code := fmt.Sprintf("<%[1]s%s>%s</%[1]s>", "code", attr, escape(n.Text))
	return wrap("pre", code)
}

func (t *Parse) newCode(pos Pos, lang, text string) *CodeNode {
	return &CodeNode{NodeType: NodeCode, Pos: pos, Lang: lang, Text: text}
}

// Link holds a tag with optional title
type LinkNode struct {
	NodeType
	Pos
	Title, Href, Text string
}

// Return the html representation of link node
func (n *LinkNode) Render() string {
	attrs := fmt.Sprintf("href=\"%s\"", n.Href)
	if n.Title != "" {
		attrs += fmt.Sprintf(" title=\"%s\"", n.Title)
	}
	return fmt.Sprintf("<a %s>%s</a>", attrs, escape(n.Text))
}

func (t *Parse) newLink(pos Pos, title, href, text string) *LinkNode {
	return &LinkNode{NodeType: NodeLink, Pos: pos, Title: title, Href: href, Text: text}
}

// RefLink holds link with refrence to link definition
type RefNode struct {
	NodeType
	Pos
	tr             *Parse
	Text, Ref, Raw string
}

// rendering based type
// TODO: Text should be TextNode(with escaping etc..)
func (n *RefNode) Render() string {
	var node Node
	ref := strings.ToLower(n.Ref)
	if l, ok := n.tr.links[ref]; ok {
		if n.Type() == NodeRefLink {
			node = n.tr.newLink(n.Pos, l.Title, l.Href, n.Text)
		} else {
			node = n.tr.newImage(n.Pos, l.Title, l.Href, n.Text)
		}
	} else {
		node = n.tr.newText(n.Pos, n.Raw)
	}
	return node.Render()
}

// create newReferenceNode(Image/Link)
func (t *Parse) newRef(typ itemType, pos Pos, raw, text, ref string) *RefNode {
	nType := NodeRefLink
	if typ == itemRefImage {
		nType = NodeRefImage
	}
	// If it's implicit link
	if ref == "" {
		ref = text
	}
	return &RefNode{NodeType: nType, Pos: pos, tr: t.root(), Raw: raw, Text: text, Ref: ref}
}

// DefLinkNode refresent single reference to link-definition
type DefLinkNode struct {
	NodeType
	Pos
	Name, Href, Title string
}

// Deflink have no representation(Transparent node)
func (n *DefLinkNode) Render() string {
	return ""
}

func (t *Parse) newDefLink(pos Pos, name, href, title string) *DefLinkNode {
	return &DefLinkNode{NodeType: NodeLink, Pos: pos, Name: name, Href: href, Title: title}
}

// Image holds img tag with optional title
type ImageNode struct {
	NodeType
	Pos
	Title, Src, Alt string
}

// Return the html representation on img node
func (n *ImageNode) Render() string {
	attrs := fmt.Sprintf("src=\"%s\" alt=\"%s\"", n.Src, n.Alt)
	if n.Title != "" {
		attrs += fmt.Sprintf(" title=\"%s\"", n.Title)
	}
	return fmt.Sprintf("<img %s>", attrs)
}

func (t *Parse) newImage(pos Pos, title, src, alt string) *ImageNode {
	return &ImageNode{NodeType: NodeImage, Pos: pos, Title: title, Src: src, Alt: alt}
}

// List holds list items nodes in ordered or unordered states.
type ListNode struct {
	NodeType
	Pos
	Ordered bool
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
		s += "\n" + item.Render()
	}
	s += "\n"
	return wrap(tag, s)
}

func (t *Parse) newList(pos Pos, ordered bool) *ListNode {
	return &ListNode{NodeType: NodeList, Pos: pos, Ordered: ordered}
}

// ListItem represent single item in ListNode that may contains nested nodes.
type ListItemNode struct {
	NodeType
	Pos
	Nodes []Node
}

func (t *ListItemNode) append(n Node) {
	t.Nodes = append(t.Nodes, n)
}

// Return the html representation of listItem
func (n *ListItemNode) Render() (s string) {
	for _, node := range n.Nodes {
		s += node.Render()
	}
	return wrap("li", s)
}

func (t *Parse) newListItem(pos Pos) *ListItemNode {
	return &ListItemNode{NodeType: NodeListItem, Pos: pos}
}

// TableNode represent table elment contains head and body
type TableNode struct {
	NodeType
	Pos
	Rows []*RowNode
}

func (t *TableNode) append(row *RowNode) {
	t.Rows = append(t.Rows, row)
}

// Return the htnml representation of a table
func (n *TableNode) Render() string {
	var s string
	for i, row := range n.Rows {
		switch i {
		case 0:
			s += wrap("thead", row.Render())
		case 1:
			s += "<tbody>"
			fallthrough
		default:
			s += row.Render()
			if i == len(n.Rows)-1 {
				s += "</tbody>"
			}
		}
	}
	return wrap("table", s)
}

func (t *Parse) newTable(pos Pos) *TableNode {
	return &TableNode{NodeType: NodeTable, Pos: pos}
}

// TableRowNode represnt tr that holds batch of table-data/cells
type RowNode struct {
	NodeType
	Pos
	Cells []*CellNode
}

func (r *RowNode) append(cell *CellNode) {
	r.Cells = append(r.Cells, cell)
}

func (r *RowNode) Render() string {
	var s string
	for _, cell := range r.Cells {
		s += cell.Render()
	}
	return wrap("tr", s)
}

func (t *Parse) newRow(pos Pos) *RowNode {
	return &RowNode{NodeType: NodeRow, Pos: pos}
}

// AlignType identifies the aligment-type of specfic cell.
type AlignType int

// Align returns itself and provides an easy default implementation
// for embedding in a Node.
func (t AlignType) Align() AlignType {
	return t
}

// Alignment
const (
	None AlignType = iota
	Right
	Left
	Center
)

// Cell types
const (
	Header = iota
	Data
)

// TableCellNode represent table-data/cell that holds simple text(may be emphasis)
// Note: the text in <th> elements are bold and centered by default.
type CellNode struct {
	NodeType
	Pos
	AlignType
	Kind  int
	Nodes []Node
}

func (c *CellNode) append(n Node) {
	c.Nodes = append(c.Nodes, n)
}

// Return the html reprenestation of table-cell
func (c *CellNode) Render() string {
	var s string
	tag := "td"
	if c.Kind == Header {
		tag = "th"
	}
	for _, node := range c.Nodes {
		s += node.Render()
	}
	return fmt.Sprintf("<%[1]s%s>%s</%[1]s>", tag, c.Style(), s)
}

// Return the cell-style based on alignment
func (c *CellNode) Style() string {
	s := " style=\"text-align:"
	switch c.Align() {
	case Right:
		s += "right\""
	case Left:
		s += "left\""
	case Center:
		s += "center\""
	default:
		s = ""
	}
	return s
}

func (t *Parse) newCell(pos Pos, kind int, align AlignType) *CellNode {
	return &CellNode{NodeType: NodeCell, Pos: pos, Kind: kind, AlignType: align}
}

// BlockQuote represent
type BlockQuoteNode struct {
	NodeType
	Pos
	Nodes []Node
}

// Render return the html representation of BlockQuote
func (n *BlockQuoteNode) Render() string {
	var s string
	for _, node := range n.Nodes {
		s += node.Render()
	}
	return wrap("blockquote", s)
}

func (t *Parse) newBlockQuote(pos Pos) *BlockQuoteNode {
	return &BlockQuoteNode{NodeType: NodeBlockQuote, Pos: pos}
}

// Wrap text with specific tag.
func wrap(tag, body string) string {
	return fmt.Sprintf("<%[1]s>%s</%[1]s>", tag, body)
}

// helper escaper
func escape(str string) (cpy string) {
	tag := regexp.MustCompile(`^<!--.*?-->|^<\/?\w+(?:"[^"]*"|'[^']*'|[^'">])*?>`)
	emp := regexp.MustCompile(`&\w+;`)
	for i := 0; i < len(str); i++ {
		switch s := str[i]; s {
		case '>':
			cpy += "&gt;"
		case '"':
			cpy += "&quot;"
		case '\'':
			cpy += "&#39;"
		case '<':
			if res := tag.FindString(str[i:]); res != "" {
				cpy += res
				i += len(res) - 1
			} else {
				cpy += "&lt;"
			}
		case '&':
			if res := emp.FindString(str[i:]); res != "" {
				cpy += res
				i += len(res) - 1
			} else {
				cpy += "&amp;"
			}
		default:
			cpy += string(s)
		}
	}
	return
}
