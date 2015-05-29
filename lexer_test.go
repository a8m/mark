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
	13: "itemLink",
	14: "itemAutoLink",
	15: "itemGfmLink",
	16: "itemStrong",
	17: "itemItalic",
	18: "itemStrike",
	19: "itemCode",
	20: "itemImage",
	21: "itemBr",
	22: "itemIndent",
}

func printRound(i int) {
	sep := strings.Repeat("#", 15)
	fmt.Printf("\n\n%s Round %d %s\n\n", sep, i, sep)
}

func TestBasic(t *testing.T) {
	l := lex("1", `
1. one
2. two
 1. one of two
 2. two of two
3. three
4. four
 1. one of four
  1. one of one of four
5. five and done!

- 1
- 2
 - 1
  - 1
  - 2
  - 3
   - 1
   - 2
   - 3
 - 2
 - 3
- 3
`)
	/*	for item := range l.items {
		fmt.Printf(tokenNames[item.typ] + " ---> '" + item.val + "'" + "\n")
	}*/
	tr := &Tree{lex: l}
	tr.parse()
	fmt.Println(tr.Nodes)
	tr.render()
	fmt.Println(tr.output)
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
