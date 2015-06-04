package mark

import (
	fmt "github.com/k0kubun/pp"
	"strings"
	"testing"
)

var tokenNames = map[itemType]string{
	-1: "itemEOF",
	0:  "itemError",
	1:  "itemNewLine",
	2:  "itemHTML",
	3:  "itemText",
	4:  "itemLineBreak",
	5:  "itemHeading",
	6:  "itemLHeading",
	7:  "itemBlockQuote",
	8:  "itemList",
	9:  "itemCodeBlock",
	10: "itemGfmCodeBlock",
	11: "itemHr",
	12: "itemTable",
	13: "itemLpTable",
	14: "itemLink",
	15: "itemAutoLink",
	16: "itemGfmLink",
	17: "itemStrong",
	18: "itemItalic",
	19: "itemStrike",
	20: "itemCode",
	21: "itemImage",
	22: "itemBr",
	23: "itemPipe",
	24: "itemIndent",
}

func printRound(i int) {
	sep := strings.Repeat("#", 15)
	fmt.Printf("\n\n%s Round %d %s\n\n", sep, i, sep)
}

func TestBasic(t *testing.T) {
	l := lex("1", `
  Name    | Age  | id
----------|------|:--
| Ariel | 26   | 2 |`)

	//	for item := range l.items {
	//		fmt.Printf(tokenNames[item.typ] + " ---> '" + item.val + "'" + "\n")
	//	}
	tr := &Tree{lex: l}
	tr.parse()
	//fmt.Println(tr.Nodes)
	//tr.render()
	//fmt.Println(tr.output)
}

func xestList(t *testing.T) {
	printRound(1)
	// Test round 1
	src := `
- foo
- bar
 - baz
1. asda
3. asdas
`
	l := lex("1", src)
	fmt.Printf("Source:\n" + src + "\n")
	for item := range l.items {
		fmt.Printf(tokenNames[item.typ] + " ---> " + item.val + "\n")
	}
}
