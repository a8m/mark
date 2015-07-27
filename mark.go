package mark

import "strings"

// Mark
type Mark struct {
	*Parse
	Input   string
	Options Options
}

// Mark options
type Options struct {
	Gfm, Tables bool
}

// Return new Mark
func New(input string) *Mark {
	// Preprocessing
	input = strings.Replace(input, "\t", "    ", -1)
	return &Mark{
		Input: input,
		Parse: newParse(input),
	}
}

// Parse and render input
func (m *Mark) Render() string {
	m.parse()
	m.render()
	return m.output
}

// AddRenderFn let you pass NodeType, and RenderFn function
// and override the default Node rendering
func (m *Mark) AddRenderFn(typ NodeType, fn RenderFn) {
	m.renderFn[typ] = fn
}

// Staic render function
func Render(input string) string {
	m := New(input)
	return m.Render()
}
