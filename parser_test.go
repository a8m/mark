package mark

import (
	"fmt"
	"github.com/k0kubun/pp"
	"testing"
)

func TestParser(t *testing.T) {
	l := lex("1", "foo bar baz\nhello world")
	p := &Tree{lex: l}
	item := p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
	p.peekCount = 0
	item = p.peek()
	fmt.Println(tokenNames[item.typ], "-->", item.val)
}

func ParseFn(*testing.T) {
	l := lex("2", `hello world
	nice to meet.`)
	p := &Tree{lex: l}
	p.parse()

	pp.Printf("[Message]: Tree Node List After Compile\n\n")
	pp.Println(p.Nodes)
}
