package mark

import (
	"fmt"
	"testing"
)

var itemName = map[itemType]string{
	eof:              "EOF",
	itemError:        "Error",
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

var lexTests = []lexTest{
	{"empty", "", []item{
		{eof, 0, ""},
	}},
	{"heading", "# Hello", []item{
		{itemHeading, 0, "# Hello"},
		{eof, 0, ""},
	}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == eof || item.typ == itemError {
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
