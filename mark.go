package mark

import "strings"

// Hook function, used for preprocessing
type HookFn func(string) string

// Mark
type Mark struct {
	Input   string
	Options Options
	Pre     []HookFn
}

// Mark options
type Options struct {
	Gfm, Tables bool
}

// Default pre function replace tabs with 4 spaces
func tabReplacer(s string) string {
	return strings.Replace(s, "\t", "    ", -1)
}

// Return new Mark
func New(input string) *Mark {
	return &Mark{
		Input: input,
		Pre:   []HookFn{tabReplacer},
	}
}

// Staic render function
func Render(input string) string {
	m := New(input)
	// PreProcessing
	for _, fn := range m.Pre {
		m.Input = fn(m.Input)
	}
	tr := &Tree{lex: lex(m.Input), links: make(map[string]*DefLinkNode)}
	tr.parse()
	tr.render()
	return tr.output
}
