package mark

import (
	"fmt"
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
	13: "itemLinks",
	14: "itemEmphasis",
	15: "itemItalic",
	16: "itemStrike",
	17: "itemCode",
	18: "itemImages",
	19: "itemBr",
}

func printRound(i int) {
	sep := strings.Repeat("#", 15)
	fmt.Printf("\n\n%s Round %d %s\n\n", sep, i, sep)
}

func TestBasic(t *testing.T) {
	printRound(1)
	// Test round 1
	l := lex("1", "#header\n#bar\n***\n---\n```js\nfunction(){}```\n~~~html\n<foo/>~~~\n##header\n```go\nmain(){}\n```")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->", item.val)
	}
	// Test round 2
	printRound(2)
	l = lex("2", "#code\n    foo bar\n    bar baz\n")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->\n"+item.val)
	}
	// Test round 3
	printRound(3)
	l = lex("3", "one  two\n\nthree  #header\n\nparagraph  ```js\ncodeblock```")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->\n"+item.val)
	}

	printRound(4)
	l = lex("4", "one ~~strike~~  two **bold**  three _italic_")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->\n"+item.val)
	}

	printRound(5)
	l = lex("5", "one ``code``  two `code`  three  ```code``  four ``code```  done")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->\n"+item.val)
	}

	printRound(6)
	l = lex("6", "- foo bar\n- baz")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->\n"+item.val)
	}

}
