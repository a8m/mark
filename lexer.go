package mark

// Token represent token or string returned from the lexer(scanner)
type Token struct {
	Type Type   // The type of this item
	Line int    // The line number on thich this token apears
	Text string // The text of this item
}

// Type identifies the type of lex items
type Type int

const (
	EOF   Type = iota // Zero value so closed channel delivers EOF
	Error             // Error occurred; value is text of error
	NewLine
	// Intersting things
	Paragraph // Simple paragraph
	Heading   // Simple heading
	LHeading  // Line heading
	Hr        // Thematic break

)
