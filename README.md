## TODO
- Inline
1. html tags
2. comments ?
3. images - (almost done)
4. escape special characters
5. mixim - **hello _world_**

- Blocks
1. heading - add id automatic.(autolink)
2. list - done!
 - break after 3-newlines(\n)
 - one or more indentation make it nested
3. table(refactor parser)
4. code should be indent with tabs too.
5. blockqoute
6. list - loose-item

- Misc
1. Escaping regex in lexer.go
2. backslash escape(inline and blocks)
3. Add peek() to lexer instead to backup all the times
4. text interpolation

Stash
-----
Some ideas:
change parseParagraph to parseInline that everyone can use it.
add ignore-list, for example ignore `br` to parseInline(for example chage `br` to to simple text)
create itemPipe and use it in the lexical phase ? or deal with it in the parseTable ?


Bugs
----
1. codeBlock as hr
`Dash

---

   ---

	    ---
 `


