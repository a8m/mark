package mark

import (
	"testing"
)

type parseTest struct {
	name  string
	items []item
	nodes []NodeType
}

type mockLexer struct {
	items []item
}

func (l *mockLexer) nextItem() (t item) {
	if len(l.items) == 0 {
		return item{itemEOF, 0, ""}
	}
	t, l.items = l.items[0], l.items[1:]
	return
}

func newMockLex(items []item) *mockLexer {
	return &mockLexer{items: items}
}

var blockparseTests = []parseTest{
	{"eof", []item{}, []NodeType{}},
	{"text-1",
		[]item{item{itemText, 0, "hello"}},
		[]NodeType{NodeParagraph},
	},
	{"text-2",
		[]item{
			item{itemText, 0, "hello"},
			item{itemNewLine, 0, "\n"},
			item{itemText, 0, "world"},
		},
		[]NodeType{NodeParagraph},
	},
	{"text-3",
		[]item{
			item{itemText, 0, "hello"},
			item{itemNewLine, 0, "\n"},
			item{itemNewLine, 0, "\n\n"},
			item{itemText, 0, "world"},
		},
		[]NodeType{NodeParagraph, NodeParagraph},
	},
	{"header",
		[]item{
			item{itemHeading, 0, "# Hello"},
		},
		[]NodeType{NodeHeading},
	},
	{"code-block",
		[]item{
			item{itemCodeBlock, 0, "    js\n    hello"},
		},
		[]NodeType{NodeCode},
	},
	{"table",
		[]item{
			item{itemTable, 0, ""},
		},
		[]NodeType{NodeTable},
	},
	{"list",
		[]item{
			item{itemList, 0, "-"},
			item{itemListItem, 0, "hello"},
		},
		[]NodeType{NodeList},
	},
	{"HTML",
		[]item{
			item{itemHTML, 0, "<hello>\nworld</hello>"},
		},
		[]NodeType{NodeHTML},
	},
}

func collectNodes(t *parseTest) []Node {
	tr := &parse{
		lex:     newMockLex(t.items),
		links:   make(map[string]*DefLinkNode),
		options: DefaultOptions(),
	}
	tr.parse()
	return tr.Nodes
}

func equalTypes(n1 []Node, n2 []NodeType) bool {
	if len(n1) != len(n2) {
		return false
	}
	for i := range n1 {
		if n1[i].Type() != n2[i] {
			return false
		}
	}
	return true
}

func TestBlocksparse(t *testing.T) {
	for _, test := range blockparseTests {
		nodes := collectNodes(&test)
		if !equalTypes(nodes, test.nodes) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, nodes, test.nodes)
		}
	}
}
