package mark

import (
	"fmt"
	"testing"
)

var itemName = map[itemType]string{
	itemError:        "Error",
	itemEOF:          "EOF",
	itemNewLine:      "NewLine",
	itemHTML:         "HTML",
	itemHeading:      "Heading",
	itemLHeading:     "LHeading",
	itemBlockQuote:   "BlockQuote",
	itemList:         "List",
	itemListItem:     "ListItem",
	itemLooseItem:    "LooseItem",
	itemCodeBlock:    "CodeBlock",
	itemGfmCodeBlock: "GfmCodeBlock",
	itemHr:           "Hr",
	itemTable:        "Table",
	itemLpTable:      "LpTable",
	itemText:         "Text",
	itemLink:         "Link",
	itemDefLink:      "DefLink",
	itemRefLink:      "RefLink",
	itemAutoLink:     "AutoLink",
	itemGfmLink:      "GfmLink",
	itemStrong:       "Strong",
	itemItalic:       "Italic",
	itemStrike:       "Strike",
	itemCode:         "Code",
	itemImage:        "Image",
	itemRefImage:     "RefImage",
	itemBr:           "Br",
	itemPipe:         "Pipe",
	itemIndent:       "Indent",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
}

var (
	tEOF = item{itemEOF, 0, ""}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"heading", "# Hello", []item{
		{itemHeading, 0, "# Hello"},
		tEOF,
	}},
	{"lheading", "Hello\n===", []item{
		{itemLHeading, 0, "Hello\n==="},
		tEOF,
	}},
	{"blockqoute", "> foo bar", []item{
		{itemBlockQuote, 0, "> foo bar"},
		tEOF,
	}},
	{"unordered list", "- foo\n- bar", []item{
		{itemList, 0, "-"},
		{itemListItem, 0, "foo"},
		{itemListItem, 0, "bar"},
		tEOF,
	}},
	{"ordered list", "1. foo\n2. bar", []item{
		{itemList, 0, "1."},
		{itemListItem, 0, "foo"},
		{itemListItem, 0, "bar"},
		tEOF,
	}},
	{"loose-items", "- foo\n\n- bar", []item{
		{itemList, 0, "-"},
		{itemLooseItem, 0, "foo"},
		{itemLooseItem, 0, "bar"},
		tEOF,
	}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, items, test.items)
		}
	}
}
