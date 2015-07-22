# Mark
> A [markdown](http://daringfireball.net/projects/markdown/) processor written in Go. built for fun.

This project ispired from [Rob Pike - Lexical Scanning talk](https://www.youtube.com/watch?v=HxaD_trXwRE) and [marked](https://github.com/chjj/marked) project.
Please note that this is a __WIP__ project and any contribution is welcomed and appreciated,
so feel free to take some task here.

## Table of contents:
- [Usage](#get-started)
- [TODO](#todo)

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


### TODO
1. Should we support def/refLink with link break?
2. backslash escape(inline and blocks)
3. Configuration options
	- gfm, table
	- heading(auto hashing)
	- smartypants
	- etc...
5. V0.2.0 - Regex is slowww, we should leave it.


### License
MIT




