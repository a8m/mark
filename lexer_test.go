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
	tEOF     = item{itemEOF, 0, ""}
	tTable   = item{itemTable, 0, ""}
	tLpTable = item{itemLpTable, 0, ""}
	tPipe    = item{itemPipe, 0, "|"}
	tNewLine = item{itemNewLine, 0, "\n"}
)

var blockTests = []lexTest{
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
	{"code-block", "    foo\n    bar", []item{
		{itemCodeBlock, 0, "    foo\n    bar"},
		tEOF,
	}},
	{"gfm-code-block-1", "~~~js\nfoo\n~~~", []item{
		{itemGfmCodeBlock, 0, "~~~js\nfoo\n~~~"},
		tEOF,
	}},
	{"gfm-code-block-2", "```js\nfoo\n```", []item{
		{itemGfmCodeBlock, 0, "```js\nfoo\n```"},
		tEOF,
	}},
	{"hr1", "* * *\n***", []item{
		{itemHr, 0, "* * *\n"},
		{itemHr, 0, "***"},
		tEOF,
	}},
	{"hr2", "- - -\n---", []item{
		{itemHr, 0, "- - -\n"},
		{itemHr, 0, "---"},
		tEOF,
	}},
	{"hr3", "_ _ _\n___", []item{
		{itemHr, 0, "_ _ _\n"},
		{itemHr, 0, "___"},
		tEOF,
	}},
	{"table", "Id | Name\n---|-----\n1 | Ariel", []item{
		tTable,
		{itemText, 0, "Id "},
		tPipe,
		{itemText, 0, " Name"},
		tNewLine,
		{itemText, 0, "---"},
		tPipe,
		{itemText, 0, "-----"},
		tNewLine,
		{itemText, 0, "1 "},
		tPipe,
		{itemText, 0, " Ariel"},
		tEOF,
	}},
	{"lp-table", "|Id | Name|\n|---|-----|\n|1 | Ariel|", []item{
		tLpTable,
		tPipe,
		{itemText, 0, "Id "},
		tPipe,
		{itemText, 0, " Name"},
		tPipe,
		tNewLine,
		tPipe,
		{itemText, 0, "---"},
		tPipe,
		{itemText, 0, "-----"},
		tPipe,
		tNewLine,
		tPipe,
		{itemText, 0, "1 "},
		tPipe,
		{itemText, 0, " Ariel"},
		tPipe,
		tEOF,
	}},
	{"text-1", "hello\nworld", []item{
		{itemText, 0, "hello"},
		tNewLine,
		{itemText, 0, "world"},
		tEOF,
	}},
	{"text-2", "__hello__\n__world__", []item{
		{itemText, 0, "__hello__"},
		tNewLine,
		{itemText, 0, "__world__"},
		tEOF,
	}},
	{"text-3", "~**_hello world_**~", []item{
		{itemText, 0, "~**_hello world_**~"},
		tEOF,
	}},
	{"text-4", "  hello world", []item{
		{itemIndent, 0, "  "},
		{itemText, 0, "hello world"},
		tEOF,
	}},
	{"deflink", "[1]: http://example.com", []item{
		{itemDefLink, 0, "[1]: http://example.com"},
		tEOF,
	}},
}

var inlineTests = []lexTest{
	{"text-1", "hello", []item{
		{itemText, 0, "hello"},
	}},
	{"text-2", "hello\nworld", []item{
		{itemText, 0, "hello\nworld"},
	}},
	{"strong-1", "**hello**", []item{
		{itemStrong, 0, "**hello**"},
	}},
	{"strong-2", "__world__", []item{
		{itemStrong, 0, "__world__"},
	}},
	{"italic-1", "*hello*", []item{
		{itemItalic, 0, "*hello*"},
	}},
	{"italic-2", "_hello_", []item{
		{itemItalic, 0, "_hello_"},
	}},
	{"strike", "~~hello~~", []item{
		{itemStrike, 0, "~~hello~~"},
	}},
	{"code", "`hello`", []item{
		{itemCode, 0, "`hello`"},
	}},
	{"link-1", "[hello](world)", []item{
		{itemLink, 0, "[hello](world)"},
	}},
	{"link-2", "[hello](world 'title')", []item{
		{itemLink, 0, "[hello](world 'title')"},
	}},
	{"autolink-1", "<http://example.com/>", []item{
		{itemAutoLink, 0, "<http://example.com/>"},
	}},
	{"autolink-2", "<http://example.com/?foo=1&bar=2>", []item{
		{itemAutoLink, 0, "<http://example.com/?foo=1&bar=2>"},
	}},
	{"gfmlink-1", "link: http://example.com/?foo=1&bar=2", []item{
		{itemText, 0, "link: "},
		{itemGfmLink, 0, "http://example.com/?foo=1&bar=2"},
	}},
	{"gfmlink-2", "http://example.com", []item{
		{itemGfmLink, 0, "http://example.com"},
	}},
	{"reflink-1", "[hello][world]", []item{
		{itemRefLink, 0, "[hello][world]"},
	}},
	{"reflink-2", "[hello]", []item{
		{itemRefLink, 0, "[hello]"},
	}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest, isInline bool) (items []item) {
	l := lex(t.input)
	if isInline {
		l = lexInline(t.input)
	}
	for item := range l.items {
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

func TestBlockLex(t *testing.T) {
	for _, test := range blockTests {
		items := collect(&test, false)
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, items, test.items)
		}
	}
}

func TestInlineLex(t *testing.T) {
	for _, test := range inlineTests {
		items := collect(&test, true)
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", test.name, items, test.items)
		}
	}
}
