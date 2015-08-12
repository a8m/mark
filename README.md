# Mark [![Test coverage][coveralls-image]][coveralls-url] [![Build status][travis-image]][travis-url]
> A [markdown](http://daringfireball.net/projects/markdown/) processor written in Go. built for fun.

This project inspired from [Rob Pike - Lexical Scanning talk](https://www.youtube.com/watch?v=HxaD_trXwRE) and [marked](https://github.com/chjj/marked) project.  
Please note that this is a WIP project and any contribution is welcomed and appreciated,
so feel free to take some task here.

## Table of contents:
- [Get Started](#get-started)
- [Examples](#examples)
- [Documentation](#documentation)
    - [Render](#render)
    - [type Mark](#mark)
        - [New](#new)
        - [AddRenderFn](#markaddrenderfn)
        - [Render](#markrender)
    - [smartypants and smartfractions](smartypants-and-smartfractions)
- [Todo](#todo)

### Get Started
#### Installation
```sh
$ go get github.com/a8m/mark
```
#### Examples
__Add to your project:__
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

__or using [mark-cli](https://github.com/a8m/mark-cli)__
```sh
$ echo 'hello __world__...' | mark-cli -smartypants
```

#### Documentation
##### Render
Staic rendering function.
```go
html := mark.Render("I am using __markdown__.")
fmt.Println(html)
// <p>I am using <strong>markdown</strong>.</p>
```

##### Mark
##### New
`New` get string as an input, and `mark.Options` as configuration and return a new `Mark`.
```go
m := mark.New("hello world...", &mark.Options{
    Smartypants: true,
})
fmt.Println(m.Render())
// <p>hello world…</p>
// Note: you can instantiate it like so: mark.New("...", nil) to get the default options.
```

##### Mark.AddRenderFn
`AddRenderFn` let you pass `NodeType`, and `RenderFn` function and override the default `Node` rendering.  
To get all Nodes type and their fields/methods, see the full documentation: [go-doc](http://godoc.org/github.com/a8m/mark)  

Example 1:
```go
m := mark.New("hello", nil)
m.AddRenderFn(mark.NodeParagraph, func(node mark.Node) (s string) {
    p, _ := node.(*mark.ParagraphNode)
    s += "<p class=\"mv-msg\">"
    for _, n := range p.Nodes {
        s += n.Render()
    }
    s += "</p>"
    return
})
fmt.Println(m.Render())
// <p class="mv-msg">hello</p>
```

Example 2:
```go
m := mark.New("# Hello world", &mark.Options{
	Smartypants: true,
	Fractions:   true,
})
m.AddRenderFn(mark.NodeHeading, func(node mark.Node) string {
	h, _ := node.(*mark.HeadingNode)
	return fmt.Sprintf("<angular-heading-directive level=\"%d\" text=\"%s\"/>", h.Level, h.Text)
})
fmt.Println(m.Render())
// <angular-heading-directive level="1" text="Hello world"/>
```

##### Mark.Render
Parse and render input.
```go
m := mark.New("hello", nil)
fmt.Println(m.Render())
// <p>hello</p>
```

#### Smartypants and Smartfractions
Mark also support [smartypants](http://daringfireball.net/projects/smartypants/) and smartfractions rendering
```go
func main() {
	opts := mark.DefaultOptions()
	opts.Smartypants = true
	opts.Fractions = true
	m := mark.New("'hello', 1/2 beer please...", opts)
	fmt.Println(m.Render())
	// ‘hello’, ½ beer please…
}
```

### Todo
- Backslash escaping
	- should ignore inside code spans
- Expand documentation
- Configuration options
	- gfm, table
	- heading(auto hashing)

### License
MIT

[travis-url]: https://travis-ci.org/a8m/mark
[travis-image]: https://img.shields.io/travis/a8m/mark.svg?style=flat-square
[coveralls-image]: https://img.shields.io/coveralls/a8m/mark.svg?style=flat-square
[coveralls-url]: https://coveralls.io/r/a8m/mark
