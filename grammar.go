package mark

import (
	"fmt"
	"regexp"
)

var (
	reGfmStart  = regexp.MustCompile("^( {0,3})([`~]{3,}) *(\\S*)?(?:.*)")
	reGfmEndGen = func(end string, i int) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`(?s)(.*?)(?:((?m)^ {0,3}%s{%d,} *$)|$)`, end, i))
	}
	reSpaceGen = func(i int) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`(?m)^ {1,%d}`, i))
	}
)
