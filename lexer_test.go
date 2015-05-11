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
	10: "itemHr",
	11: "itemTable",
	12: "itemLinks",
	13: "itemEmphasis",
	14: "itemItalic",
	15: "itemStrike",
	16: "itemCode",
	17: "itemImages",
}

func TestBasic(t *testing.T) {
	l := lex("foo", "#headder \n#bar \n*** \n---")
	for item := range l.items {
		fmt.Println(tokenNames[item.typ], "--->", item.val)
	}
}
