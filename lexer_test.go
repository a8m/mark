package mark

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	l := lex("foo", "#headder \n#bar")
	for item := range l.items {
		fmt.Println(item)
	}
}
