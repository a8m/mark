package mark

import (
	"fmt"
	"testing"
)

var tokenNames = map[itemType]string{
	0:  "itemError",
	1:  "itemNewLine",
	2:  "itemHTML",
	3:  "itemParagraph",
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
}

func TestBasic(t *testing.T) {
	// Test round 1
	l := lex("foo", "#header\n#bar\n***\n---\n```js\nfunction(){}```\n~~~html\n<foo/>~~~\n##header\n```go\nmain(){}\n```")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->", item.val)
	}
	// Test round 2

}
