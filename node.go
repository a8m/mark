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
	NodeList
)

// ParagraphNode hold simple paragraph node contains text
// that may be emphasis.
type ParagraphNode struct {
	NodeType
	Nodes []Node
	// tr NodeTree
}

// TextNode holds plain text.
type TextNode struct {
	NodeType
	text []byte
	// tr NodeTree
}
