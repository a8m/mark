# TODO
- Inline
1. href (almost done, works only when it's on the start of the line)
2. html tags
3. comments ?
4. images - (almost done)
5. itemIndent

- Blocks
1. heading - add id automatic.(autolink)
2. list
 - break after 3-newlines(\n)
 - one or more indentation make it nested
3. table
4. code should be indent with tabs too.

- Misc
1. Escaping regex in lexer.go
2. backslash escape(inline and blocks)
3. Add peek() to lexer instead to backup all the times
4. text interpolation

- Preprocessing
src = src
    .replace(/\r\n|\r/g, '\n')
    .replace(/\t/g, '    ')
    .replace(/\u00a0/g, ' ')
    .replace(/\u2424/g, '\n');

Stash
-----
	re := regexp.MustCompile(`^\[((?:\[[^\]]*\]|[^\[\]]|\])*)\]\(\s*<?([\s\S]*?)>?(?:\s+['"]([\s\S]*?)['"])?\s*\)`)
	fmt.Println(re.FindStringSubmatch("[name](link \"title\")"))
	fmt.Println(re.FindStringSubmatch("[name]()"))
	fmt.Println(re.FindStringSubmatch("[name](link)"))
	fmt.Println(re.FindStringSubmatch("[name](http://adasnd.com?asdas3 \"Title\")"))
	
	re = regexp.MustCompile(`^<([^ >]+(@|:\/)[^ >]+)>`)
	// Autolinks
	// <http://example.com/>
	fmt.Println(re.FindStringSubmatch("<http://lol.com>"))
	
	re = regexp.MustCompile(`^(https?:\/\/[^\s<]+[^<.,:;"')\]\s])`)
	// Gfm links
	// This should be a link: http://example.com/hello-world.
	fmt.Println(re.FindStringSubmatch("http://asdasdsa.com"))
