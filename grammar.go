package mark

import (
	"fmt"
	"regexp"
)

var (
	reHr         = regexp.MustCompile(`^(?:(?:\* *){3,}|(?:_ *){3,}|(?:- *){3,}) *(?:\n+|$)`)
	reHeading    = regexp.MustCompile(`^ *(#{1,6}) +([^\n]+?) *#* *(?:\n|$)`)
	reLHeading   = regexp.MustCompile(`^([^\n]+?) *\n {0,3}(=|-){1,} *(?:\n+|$)`)
	reBlockQuote = regexp.MustCompile(`^( *>[^\n]*(\n[^\n]+)*\n*)+`)
	reSpaceGen   = func(i int) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`(?m)^ {1,%d}`, i))
	}
)

var reCodeBlock = struct {
	*regexp.Regexp
	trim func(src, repl string) string
}{
	regexp.MustCompile(`^( {4}[^\n]+(?: *\n)*)+`),
	regexp.MustCompile("(?m)^( {0,4})").ReplaceAllLiteralString,
}

var reGfmCode = struct {
	*regexp.Regexp
	endGen func(end string, i int) *regexp.Regexp
}{
	regexp.MustCompile("^( {0,3})([`~]{3,}) *(\\S*)?(?:.*)"),
	func(end string, i int) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`(?s)(.*?)(?:((?m)^ {0,3}%s{%d,} *$)|$)`, end, i))
	},
}

var reTable = struct {
	item, itemLp *regexp.Regexp
	split        func(s string, n int) []string
	trim         func(src, repl string) string
}{
	regexp.MustCompile(`^ *(\S.*\|.*)\n *([-:]+ *\|[-| :]*)\n((?:.*\|.*(?:\n|$))*)\n*`),
	regexp.MustCompile(`(^ *\|.+)\n( *\| *[-:]+[-| :]*)\n((?: *\|.*(?:\n|$))*)\n*`),
	regexp.MustCompile(` *\| *`).Split,
	regexp.MustCompile(`^ *\| *| *\| *$`).ReplaceAllString,
}

var reHTML = struct {
	item, comment, tag, span *regexp.Regexp
	endTagGen                func(tag string) *regexp.Regexp
}{
	regexp.MustCompile(`^<(\w+)(?:"[^"]*"|'[^']*'|[^'">])*?>`),
	regexp.MustCompile(`(?sm)<!--.*?-->`),
	regexp.MustCompile(`^<!--.*?-->|^<\/?\w+(?:"[^"]*"|'[^']*'|[^'">])*?>`),
	// TODO: Add all span-tags and move to config.
	regexp.MustCompile(`^(a|em|strong|small|s|q|data|time|code|sub|sup|i|b|u|span|br|del|img)$`),
	func(tag string) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`(?s)(.)+?<\/%s> *(?:\n{2,}|\s*$)`, tag))
	},
}
