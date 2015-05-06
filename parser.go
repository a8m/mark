package mark

type Tree struct {
	text string
	lex  *lexer
}

func Parse(text string) (err error) {
	// Create TreeSet
	t := &Tree{text, new(lexer)}
	err = t.Parse()
	return
}

func (t *Tree) Parse() error {
	t.parse()
	return nil
}

func (t *Tree) parse() {

}
