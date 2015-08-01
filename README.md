# Mark [![Test coverage][coveralls-image]][coveralls-url] [![Build status][travis-image]][travis-url]
> A [markdown](http://daringfireball.net/projects/markdown/) processor written in Go. built for fun.

This project inspired from [Rob Pike - Lexical Scanning talk](https://www.youtube.com/watch?v=HxaD_trXwRE) and [marked](https://github.com/chjj/marked) project.  
Please note that this is a __WIP__ project and any contribution is welcomed and appreciated,
so feel free to take some task here.

## Table of contents:
- [Usage](#usage)
- [Todo](#todo)

### Usage
#### Installation
```sh
$ go get github.com/a8m/mark
```
#### Add to your project
```go
import (
	"fmt"
	"github.com/a8m/mark"
)

func main() {
	html := mark.Render("I am using __markdown__.")
	fmt.Println(html)
	// <p>I am using <strong>markdown</strong>.</p>
}
```
#### Override default rendering
**Usage:** `m.AddRenderFn(NodeType, func(Node) string)`
```go
func main() {
	m := mark.New("hello")
	m.AddRenderFn(mark.NodeParagraph, func(n mark.Node) (s string) {
		if p, ok := n.(*mark.ParagraphNode); ok {
			s += "<p class=\"mv-msg\">"
			for _, node := range p.Nodes {
				s += node.Render()
			}
			s += "</p>"
		}
		return
	})
	fmt.Println(m.Render())
	// <p class="mv-msg">hello</p>
}
```


### Todo
- Backslash escape
- Expand documation
- Configuration options
	- gfm, table
	- heading(auto hashing)
	- smartypants
	- etc...

### License
MIT

[travis-url]: https://travis-ci.org/a8m/mark
[travis-image]: https://img.shields.io/travis/a8m/mark.svg?style=flat-square
[coveralls-image]: https://img.shields.io/coveralls/a8m/mark.svg?style=flat-square
[coveralls-url]: https://coveralls.io/r/a8m/mark

