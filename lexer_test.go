package mark

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	l := lex("foo", "#head \ndsadasd")
	fmt.Println(<-l.items)
}
