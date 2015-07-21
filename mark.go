package mark

import "strings"

// Just for test right now
func Render(text string) string {
	// Preproessing
	// 1. Replace all tabs with 4-spaces
	text = strings.Replace(text, "\t", "    ", -1)
	// TODO: Use/ot remove name option
	t := &Tree{lex: lex(text), links: make(map[string]*DefLinkNode)}
	t.parse()
	// PostProcessing
	// 1. HTML escaping(<, >, ...)
	t.render()
	return t.output
}
