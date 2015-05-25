package mark

// Just for test right now
func Render(text string) string {
	// TODO: Use name option
	t := &Tree{lex: lex(text, text)}
	t.parse()
	t.render()
	return t.output
}
