package mark

// A Node is an element in the parse tree.
type Node interface {
	Type() NodeType
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
	NodeNewLine
	NodeList
)

// ParagraphNode hold simple paragraph node contains text
// that may be emphasis.
type ParagraphNode struct {
	NodeType
	Pos
	Nodes []Node
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

func (t *Tree) newText(pos Pos, text string) *TextNode {
	return &TextNode{NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

// NewLineNode represent simple `\n`
type NewLineNode struct {
	NodeType
	Pos
}

func (t *Tree) newLine(pos Pos) *NewLineNode {
	return &NewLineNode{NodeType: NodeNewLine, Pos: pos}
}
