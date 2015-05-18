package mark

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	l := lex("1", "foo bar baz\n")
	p := &Tree{lex: l}
	item := p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
	p.peekCount = 0
	item = p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
}

func TestParseFn(*testing.T) {
	l := lex("2", "hello world")
	p := &Tree{lex: l}
	p.parse()
}
