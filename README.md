## TODO
- Inline
1. html tags
3. images - (almost done - defined image reference)
  - link reference
4. escape special characters
5. refLink will bind to defLink in the rendering phase

- Blocks
3. table(refactor parser)
4. code should be indent with tabs too.(fix with preprocessing)
5. blockqoute(ignore defLink)

- Misc
1. Escaping regex in lexer.go
2. backslash escape(inline and blocks)
3. Add peek() to lexer instead to backup all the times
4. text interpolation
5. Configuration, gfm or not(heading, tables, spanTags, etc...)

Stash
-----
Some ideas:
change parseParagraph to parseInline that everyone can use it.
add ignore-list, for example ignore `br` to parseInline(for example chage `br` to to simple text)
create itemPipe and use it in the lexical phase ? or deal with it in the parseTable ?




