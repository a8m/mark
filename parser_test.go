package mark

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	l := lex("1", "foo bar baz")
	p := &Tree{lex: l}
	item := p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
	p.peekCount = 0
	item = p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
}
