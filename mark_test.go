package mark

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	cases := map[string]string{
		"foobar":               "<p>foobar</p>",
		"  foo bar":            "<p>foo bar</p>",
		"|foo|bar":             "<p>|foo|bar</p>",
		"foo  \nbar":           "<p>foo<br>bar</p>",
		"__bar__ foo":          "<p><strong>bar</strong> foo</p>",
		"**bar** foo __bar__":  "<p><strong>bar</strong> foo <strong>bar</strong></p>",
		"**bar**__baz__":       "<p><strong>bar</strong><strong>baz</strong></p>",
		"**bar**foo__bar__":    "<p><strong>bar</strong>foo<strong>bar</strong></p>",
		"_bar_baz":             "<p><em>bar</em>baz</p>",
		"_foo_~~bar~~ baz":     "<p><em>foo</em><del>bar</del> baz</p>",
		"~~baz~~ _baz_":        "<p><del>baz</del> <em>baz</em></p>",
		"`bool` and thats it.": "<p><code>bool</code> and thats it.</p>",
		// Html
		"<!--hello-->": "<!--hello-->",
		// Emphasis mixim
		"___foo___":       "<p><strong><em>foo</em></strong></p>",
		"__foo _bar___":   "<p><strong>foo <em>bar</em></strong></p>",
		"__*foo*__":       "<p><strong><em>foo</em></strong></p>",
		"_**mixim**_":     "<p><em><strong>mixim</strong></em></p>",
		"~~__*mixim*__~~": "<p><del><strong><em>mixim</em></strong></del></p>",
		"~~*mixim*~~":     "<p><del><em>mixim</em></del></p>",
		// Paragraph
		"1  \n2  \n3":        "<p>1<br>2<br>3</p>",
		"1\n\n2":             "<p>1</p>\n<p>2</p>",
		"1\n\n\n2":           "<p>1</p>\n<p>2</p>",
		"1\n\n\n\n\n\n\n\n2": "<p>1</p>\n<p>2</p>",
		// Heading
		"# 1\n## 2":                   "<h1 id=\"1\">1</h1>\n<h2 id=\"2\">2</h2>",
		"# 1\np\n## 2\n### 3\n4\n===": "<h1 id=\"1\">1</h1>\n<p>p</p>\n<h2 id=\"2\">2</h2>\n<h3 id=\"3\">3</h3>\n<h1 id=\"4\">4</h1>",
		"Hello\n===":                  "<h1 id=\"hello\">Hello</h1>",
		// Links
		"[text](link \"title\")": "<p><a href=\"link\" title=\"title\">text</a></p>",
		"[text](link)":           "<p><a href=\"link\">text</a></p>",
		"[](link)":               "<p><a href=\"link\"></a></p>",
		"Link: [example](#)":     "<p>Link: <a href=\"#\">example</a></p>",
		"Link: [not really":      "<p>Link: [not really</p>",
		"http://localhost:3000":  "<p><a href=\"http://localhost:3000\">http://localhost:3000</a></p>",
		"Link: http://yeah.com":  "<p>Link: <a href=\"http://yeah.com\">http://yeah.com</a></p>",
		"<http://foo.com>":       "<p><a href=\"http://foo.com\">http://foo.com</a></p>",
		"Link: <http://l.co>":    "<p>Link: <a href=\"http://l.co\">http://l.co</a></p>",
		"Link: <not really":      "<p>Link: &lt;not really</p>",
		// CodeBlock
		"\tfoo\n\tbar": "<pre><code>foo\nbar</code></pre>",
		"\tfoo\nbar":   "<pre><code>foo\n</code></pre>\n<p>bar</p>",
		// GfmCodeBlock
		"```js\nvar a;\n```":         "<pre><code class=\"lang-js\">\nvar a;\n</code></pre>",
		"~~~\nvar b;~~let d = 1~~~~": "<pre><code>\nvar b;~~let d = 1~~~~</code></pre>",
		"~~~js\n":                    "<pre><code class=\"lang-js\">\n</code></pre>",
		// Hr
		"foo\n****\nbar": "<p>foo</p>\n<hr>\n<p>bar</p>",
		"foo\n___":       "<p>foo</p>\n<hr>",
		// Images
		"![name](url)":           "<p><img src=\"url\" alt=\"name\"></p>",
		"![name](url \"title\")": "<p><img src=\"url\" alt=\"name\" title=\"title\"></p>",
		"img: ![name]()":         "<p>img: <img src=\"\" alt=\"name\"></p>",
		// Lists
		"- foo\n- bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		"* foo\n* bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		"+ foo\n+ bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		// // Ordered Lists
		"1. one\n2. two\n3. three": "<ol>\n<li>one</li>\n<li>two</li>\n<li>three</li>\n</ol>",
		"1. one\n 1. one of one":   "<ol>\n<li>one<ol>\n<li>one of one</li>\n</ol></li>\n</ol>",
		"2. two\n 3. three":        "<ol>\n<li>two<ol>\n<li>three</li>\n</ol></li>\n</ol>",
		// Task list
		"- [ ] foo\n- [ ] bar": "<ul>\n<li><input type=\"checkbox\">foo</li>\n<li><input type=\"checkbox\">bar</li>\n</ul>",
		"- [x] foo\n- [x] bar": "<ul>\n<li><input type=\"checkbox\" checked>foo</li>\n<li><input type=\"checkbox\" checked>bar</li>\n</ul>",
		"- [ ] foo\n- [x] bar": "<ul>\n<li><input type=\"checkbox\">foo</li>\n<li><input type=\"checkbox\" checked>bar</li>\n</ul>",
		// Special characters escaping
		"< hello":   "<p>&lt; hello</p>",
		"hello >":   "<p>hello &gt;</p>",
		"foo & bar": "<p>foo &amp; bar</p>",
		"'foo'":     "<p>&#39;foo&#39;</p>",
		"\"foo\"":   "<p>&quot;foo&quot;</p>",
		"&copy;":    "<p>&copy;</p>",
		// Backslash escaping
		"\\**foo\\**":       "<p>*<em>foo*</em></p>",
		"\\*foo\\*":         "<p>*foo*</p>",
		"\\_underscores\\_": "<p>_underscores_</p>",
		"\\## header":       "<p>## header</p>",
		"header\n\\===":     "<p>header\n\\===</p>",
	}
	for input, expected := range cases {
		if actual := Render(input); actual != expected {
			t.Errorf("%s: got\n%+v\nexpected\n%+v", input, actual, expected)
		}
	}
}

func TestData(t *testing.T) {
	var testFiles []string
	files, err := ioutil.ReadDir("test")
	if err != nil {
		t.Error("Couldn't open 'test' directory")
	}
	for _, file := range files {
		if name := file.Name(); strings.HasSuffix(name, ".text") {
			testFiles = append(testFiles, "test/"+strings.TrimSuffix(name, ".text"))
		}
	}
	re := regexp.MustCompile(`\n`)
	for _, file := range testFiles {
		html, err := ioutil.ReadFile(file + ".html")
		if err != nil {
			t.Errorf("Error to read html file: %s", file)
		}
		text, err := ioutil.ReadFile(file + ".text")
		if err != nil {
			t.Errorf("Error to read text file: %s", file)
		}
		// Remove '\n'
		sHTML := re.ReplaceAllLiteralString(string(html), "")
		output := Render(string(text))
		opts := DefaultOptions()
		if strings.Contains(file, "smartypants") {
			opts.Smartypants = true
			output = New(string(text), opts).Render()
		}
		if strings.Contains(file, "smartyfractions") {
			opts.Fractions = true
			output = New(string(text), opts).Render()
		}
		sText := re.ReplaceAllLiteralString(output, "")
		if sHTML != sText {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", file, sText, sHTML)
		}
	}
}

// TODO: Add more tests for it.
func TestRenderFn(t *testing.T) {
	m := New("hello world", nil)
	m.AddRenderFn(NodeParagraph, func(n Node) (s string) {
		if p, ok := n.(*ParagraphNode); ok {
			s += "<p class=\"mv-msg\">"
			for _, pp := range p.Nodes {
				s += pp.Render()
			}
			s += "</p>"
		}
		return
	})
	expected := "<p class=\"mv-msg\">hello world</p>"
	if actual := m.Render(); actual != expected {
		t.Errorf("RenderFn: got\n\t%+v\nexpected\n\t%+v", actual, expected)
	}
}

type CommonMarkSpec struct {
	name     string
	input    string
	expected string
}

var CMCases = []CommonMarkSpec{
	{"6", "- `one\n- two`", "<ul><li>`one</li><li>two`</li></ul>"},
	{"7", "***\n---\n___", "<hr><hr><hr>"},
	{"8", "+++", "<p>+++</p>"},
	{"9", "===", "<p>===</p>"},
	{"10", "--\n**\n__", "<p>--**__</p>"},
	{"11", " ***\n  ***\n   ***", "<hr><hr><hr>"},
	{"12", "    ***", "<pre><code>***</code></pre>"},
	{"14", "_____________________________________", "<hr>"},
	{"15", " - - -", "<hr>"},
	{"16", " **  * ** * ** * **", "<hr>"},
	{"17", "-     -      -      -", "<hr>"},
	{"18", "- - - -    ", "<hr>"},
	{"20", " *-*", "<p><em>-</em></p>"},
	{"21", "- foo\n***\n- bar", "<ul>\n<li>foo</li>\n</ul>\n<hr>\n<ul>\n<li>bar</li>\n</ul>"},
	{"22", "Foo\n***\nbar", "<p>Foo</p><hr><p>bar</p>"},
	{"23", "Foo\n---\nbar", "<h2>Foo</h2><p>bar</p>"},
	{"24", "* Foo\n* * *\n* Bar", "<ul>\n<li>Foo</li>\n</ul>\n<hr>\n<ul>\n<li>Bar</li>\n</ul>"},
	{"25", "- Foo\n- * * *", "<ul>\n<li>Foo</li>\n<li>\n<hr>\n</li>\n</ul>"},
	{"26", `# foo
## foo
### foo
#### foo
##### foo
###### foo`, `<h1>foo</h1>
<h2>foo</h2>
<h3>foo</h3>
<h4>foo</h4>
<h5>foo</h5>
<h6>foo</h6>`},
	{"27", "####### foo", "<p>####### foo</p>"},
	{"28", "#5 bolt\n\n#foobar", "<p>#5 bolt</p>\n<p>#foobar</p>"},
	{"29", "\\## foo", "<p>## foo</p>"},
	{"30", "# foo *bar* \\*baz\\*", "<h1>foo <em>bar</em> *baz*</h1>"},
	{"31", "#                  foo                     ", "<h1>foo</h1>"},
	{"32", ` ### foo
  ## foo
   # foo`, `<h3>foo</h3>
<h2>foo</h2>
<h1>foo</h1>`},
	{"33", "    # foo", "<pre><code># foo</code></pre>"},
	{"34", `
foo
    # bar`, `
<p>foo
# bar</p>`},
	{"35", `## foo ##
  ###   bar    ###`, `<h2>foo</h2>
<h3>bar</h3>`},
	{"36", `# foo ##################################
##### foo ##`, `<h1>foo</h1>
<h5>foo</h5>`},
	{"37", "### foo ###     ", "<h3>foo</h3>"},
	{"38", "### foo ### b", "<h3>foo ### b</h3>"},
	{"39", "# foo#", "<h1>foo#</h1>"},
	{"40", `
### foo \###
## foo #\##
# foo \#`, `
<h3>foo ###</h3>
<h2>foo ###</h2>
<h1>foo #</h1>`},
	{"41", `****
## foo
****`, `<hr>
<h2>foo</h2>
<hr>`},
	{"42", `Foo bar
# baz
Bar foo`, `<p>Foo bar</p>
<h1>baz</h1>
<p>Bar foo</p>`},
	{"43", `
## 
#
### ###`, `
<h2></h2>
<h1></h1>
<h3></h3>`},
	{"44", `
Foo *bar*
=========

Foo *bar*
---------`, `
<h1>Foo <em>bar</em></h1>
<h2>Foo <em>bar</em></h2>`},
	{"45", `Foo
-------------------------

Foo
=`, `<h2>Foo</h2>
<h1>Foo</h1>`},
	{"46", `   Foo
---

  Foo
-----

  Foo
  ===`, `<h2>Foo</h2>
<h2>Foo</h2>
<h1>Foo</h1>`},
	{"47", `    Foo
    ---

    Foo
---`, `<pre><code>Foo
---

Foo
</code></pre>
<hr>`},
	{"48", `Foo
   ----      `, "<h2>Foo</h2>"},
	{"49", `
 Foo
    ---`, `
<p>Foo
---</p>`},
	{"50", `Foo
= =

Foo
--- -`, `<p>Foo
= =</p>
<p>Foo</p>
<hr>`},
	{"51", `Foo  
-----`, "<h2>Foo</h2>"},
	{"52", `Foo\
----`, "<h2>Foo\\</h2>"},
	{"53", "`Foo\n----\n`\n\n<a title=\"a lot\n---\nof dashes\"/>", "<h2>`Foo</h2>\n<p>`</p>\n<h2>&lt;a title=&quot;a lot</h2>\n<p>of dashes&quot;/&gt;</p>"},
	{"54", `
> Foo
---`, `
<blockquote>
<p>Foo</p>
</blockquote>
<hr>`},
	{"55", `- Foo
---`, `<ul>
<li>Foo</li>
</ul>
<hr>`},
	{"57", `---
Foo
---
Bar
---
Baz`, `<hr>
<h2>Foo</h2>
<h2>Bar</h2>
<p>Baz</p>`},
	{"58", "====", "<p>====</p>"},
	{"59", `---
---`, "<hr><hr>"},
	{"60", `- foo
-----`, `<ul>
<li>foo</li>
</ul>
<hr>`},
	{"61", `    foo
---`, `<pre><code>foo
</code></pre>
<hr>`},
	{"62", `
> foo
-----`, `
<blockquote>
<p>foo</p>
</blockquote>
<hr>`},
	{"63", `
\> foo
------`, `
<h2>&gt; foo</h2>`},
	{"64", `    a simple
      indented code block`, `<pre><code>a simple
  indented code block
</code></pre>`},
	{"65", `
  - foo

    bar`, `
<ul>
<li>
<p>foo</p>
<p>bar</p>
</li>
</ul>`},
	{"66", `1.  foo

    - bar`, `<ol>
<li>
<p>foo</p>
<ul>
<li>bar</li>
</ul>
</li>
</ol>`},
	{"67", `    <a/>
    *hi*

    - one`, `<pre><code>&lt;a/&gt;
*hi*

- one
</code></pre>`},
	{"68", `
    chunk1

    chunk2
  
 
 
    chunk3`, `
<pre><code>chunk1

chunk2



chunk3
</code></pre>`},
	{"69", `
    chunk1
      
      chunk2`, `
<pre><code>chunk1
  
  chunk2
</code></pre>`},
	{"70", `
Foo
    bar`, `
<p>Foo
bar</p>`},
	{"71", `    foo
bar`, `<pre><code>foo
</code></pre>
<p>bar</p>`},
	{"72", `# Header
    foo
Header
------
    foo
----`, `<h1>Header</h1>
<pre><code>foo
</code></pre>
<h2>Header</h2>
<pre><code>foo
</code></pre>
<hr>`},
	{"73", `        foo
    bar`, `<pre><code>    foo
bar
</code></pre>`},
	{"74", `    
    foo
    `, `<pre><code>foo
</code></pre>`},
	{"75", "    foo  ", `<pre><code>foo  
</code></pre>`},
	{"76", "```\n< \n>\n```", `<pre><code>&lt;
 &gt;
</code></pre>`},
	{"77", `~~~
<
 >
~~~`, `<pre><code>&lt;
 &gt;
</code></pre>`},
	{"78", "```\naaa\n~~~\n```", `<pre><code>aaa
~~~
</code></pre>`},
	{"79", "~~~\naaa\n```\n~~~", "<pre><code>aaa\n```\n</code></pre>"},
	{"80", "````\naaa\n```\n``````", "<pre><code>aaa\n```\n</code></pre>"},
	{"81", `
~~~~
aaa
~~~
~~~~`, `
<pre><code>aaa
~~~
</code></pre>`},
	{"82", "```", "<pre><code></code></pre>"},
	{"83", "`````\n\n```\naaa", "<pre><code>\n```\naaa\n</code></pre>"},
	{"84", "> ```\n> aaa\n\nbbb", `
<blockquote>
<pre><code>aaa
</code></pre>
</blockquote>
<p>bbb</p>`},
	{"85", "```\n\n  \n```", "<pre><code>\n  \n</code></pre>"},
	{"86", "```\n```", `<pre><code></code></pre>`},
	{"87", " ```\n aaa\naaa\n```", `
<pre><code>aaa
aaa
</code></pre>`},
	{"88", "  ```\naaa\n  aaa\naaa\n  ```", `
<pre><code>aaa
aaa
aaa
</code></pre>`},
	{"89", "   ```\n   aaa\n    aaa\n  aaa\n   ```", `
<pre><code>aaa
 aaa
aaa
</code></pre>`},
	{"90", "    ```\n    aaa\n    ```", "<pre><code>```\naaa\n```\n</code></pre>"},
	{"91", "```\naaa\n  ```", `<pre><code>aaa
</code></pre>`},
	{"92", "   ```\naaa\n  ```", `<pre><code>aaa
</code></pre>`},
	{"93", "```\naaa\n    ```", "<pre><code>aaa\n    ```\n</code></pre>"},
	{"95", `
~~~~~~
aaa
~~~ ~~`, `
<pre><code>aaa
~~~ ~~
</code></pre>`},
	{"96", "foo\n```\nbar\n```\nbaz", `<p>foo</p>
<pre><code>bar
</code></pre>
<p>baz</p>`},
	{"97", `foo
---
~~~
bar
~~~
# baz`, `<h2>foo</h2>
<pre><code>bar
</code></pre>
<h1>baz</h1>`},
	{"102", "```\n``` aaa\n```", "<pre><code>``` aaa\n</code></pre>"},
	{"103", `
<table>
  <tr>
    <td>
           hi
    </td>
  </tr>
</table>

okay.`, `
<table>
  <tr>
    <td>
           hi
    </td>
  </tr>
</table>
<p>okay.</p>`},
	// Move out the id, beacuse the regexp below
	{"107", `
<div
  class="bar">
</div>`, `
<div
  class="bar">
</div>`},
	{"108", `
<div class="bar
  baz">
</div>`, `
<div class="bar
  baz">
</div>`},
	{"113", `<div><a href="bar">*foo*</a></div>`, `<div><a href="bar">*foo*</a></div>`},
	{"114", `
<table><tr><td>
foo
</td></tr></table>`, `
<table><tr><td>
foo
</td></tr></table>`},
	{"117", `
<Warning>
*bar*
</Warning>`, `
<Warning>
*bar*
</Warning>`},
	{"121", "<del>*foo*</del>", "<p><del><em>foo</em></del></p>"},
	{"122", `
<pre language="haskell"><code>
import Text.HTML.TagSoup

main :: IO ()
main = print $ parseTags tags
</code></pre>`, `
<pre language="haskell"><code>
import Text.HTML.TagSoup

main :: IO ()
main = print $ parseTags tags
</code></pre>`},
	{"123", `
<script type="text/javascript">
// JavaScript example

document.getElementById("demo").innerHTML = "Hello JavaScript!";
</script>`, `
<script type="text/javascript">
// JavaScript example

document.getElementById("demo").innerHTML = "Hello JavaScript!";
</script>`},
	{"124", `
<style
  type="text/css">
h1 {color:red;}

p {color:blue;}
</style>`, `
<style
  type="text/css">
h1 {color:red;}

p {color:blue;}
</style>`},
	{"127", `
- <div>
- foo`, `
<ul>
<li>
<div>
</li>
<li>foo</li>
</ul>`},
	{"137", `
Foo
<div>
bar
</div>`, `
<p>Foo</p>
<div>
bar
</div>`},
	{"139", `
Foo
<a href="bar">
baz`, `
<p>Foo
<a href="bar">
baz</p>`},
	{"141", `
<div>
*Emphasized* text.
</div>`, `
<div>
*Emphasized* text.
</div>
`},
	{"142", `
<table>

<tr>

<td>
Hi
</td>

</tr>

</table>`, `
<table>
<tr>
<td>
Hi
</td>
</tr>
</table>
`},
	{"144", `
[foo]: /url "title"

[foo]`, `<p><a href="/url" title="title">foo</a></p>`},
	{"145", `
   [foo]: 
      /url  
           'the title'  

[foo]`, `<p><a href="/url" title="the title">foo</a></p>`},
	{"148", `
[foo]: /url '
title
line1
line2
'

[foo]`, `
<p><a href="/url" title="
title
line1
line2
">foo</a></p>`},
	{"150", `
[foo]:
/url

[foo]`, `<p><a href="/url">foo</a></p>`},
	{"151", `
[foo]:

[foo]`, `
<p>[foo]:</p>
<p>[foo]</p>`},
	{"153", `
[foo]

[foo]: url`, `<p><a href="url">foo</a></p>`},
	{"154", `
[foo]

[foo]: first
[foo]: second`, `<p><a href="first">foo</a></p>`},
	{"155", `
[FOO]: /url

[Foo]`, `<p><a href="/url">Foo</a></p>`},
	{"157", "[foo]: /url", ""},
	{"158", `
[
foo
]: /url
bar`, "<p>bar</p>"},
	{"159", `[foo]: /url "title" ok`, "<p>[foo]: /url &quot;title&quot; ok</p>"},
	{"160", `
[foo]: /url
"title" ok`, "<p>&quot;title&quot; ok</p>"},
	{"161", `
    [foo]: /url "title"

[foo]`, `
<pre><code>[foo]: /url &quot;title&quot;
</code></pre>
<p>[foo]</p>`},
	{"162", "```\n[foo]: /url\n```\n\n[foo]", `
<pre><code>[foo]: /url
</code></pre>
<p>[foo]</p>`},
	{"166", `
[foo]

> [foo]: /url`, `
<p><a href="/url">foo</a></p>
<blockquote>
</blockquote>`},
	{"167", `
aaa

bbb`, `
<p>aaa</p>
<p>bbb</p>`},
	{"168", `
aaa
bbb

ccc
ddd`, `
<p>aaa
bbb</p>
<p>ccc
ddd</p>`},
	{"169", `
aaa


bbb`, `
<p>aaa</p>
<p>bbb</p>`},
	{"170", `
  aaa
 bbb`, `
<p>aaa
bbb</p>`},
	{"171", `
aaa
             bbb
                                       ccc`, `
<p>aaa
bbb
ccc</p>`},
	{"172", `
   aaa
bbb`, `
<p>aaa
bbb</p>`},
	{"173", `
    aaa
bbb`, `
<pre><code>aaa
</code></pre>
<p>bbb</p>`},
	{"174", `
aaa     
bbb     `, `
<p>aaa<br>
bbb</p>`},
	{"175", `
  

aaa
  

# aaa

  `, `
<p>aaa</p>
<h1>aaa</h1>`},
	{"176", `
> # Foo
> bar
> baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"177", `
># Foo
>bar
> baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"178", `
   > # Foo
   > bar
 > baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"179", `
    > # Foo
    > bar
    > baz`, `
<pre><code>&gt; # Foo
&gt; bar
&gt; baz
</code></pre>`},
	{"180", `
> # Foo
> bar
baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"181", `
> bar
baz
> foo`, `
<blockquote>
<p>bar
baz
foo</p>
</blockquote>`},
	{"182", `
> foo
---`, `
<blockquote>
<p>foo</p>
</blockquote>
<hr>`},
	{"186", `
>`, `
<blockquote>
</blockquote>`},
	{"187", `
>
>  
> `, `
<blockquote>
</blockquote>`},
	{"188", `
>
> foo
>  `, `
<blockquote>
<p>foo</p>
</blockquote>`},
	{"189", `
> foo

> bar`, `
<blockquote>
<p>foo</p>
</blockquote>
<blockquote>
<p>bar</p>
</blockquote>`},
	{"190", `
> foo
> bar`, `
<blockquote>
<p>foo
bar</p>
</blockquote>`},
	{"191", `
> foo
>
> bar`, `
<blockquote>
<p>foo</p>
<p>bar</p>
</blockquote>`},
	{"192", `
foo
> bar`, `
<p>foo</p>
<blockquote>
<p>bar</p>
</blockquote>`},
	{"193", `
> aaa
***
> bbb`, `
<blockquote>
<p>aaa</p>
</blockquote>
<hr>
<blockquote>
<p>bbb</p>
</blockquote>`},
	{"194", `
> bar
baz`, `
<blockquote>
<p>bar
baz</p>
</blockquote>`},
	{"195", `
> bar

baz`, `
<blockquote>
<p>bar</p>
</blockquote>
<p>baz</p>`},
	{"197", `
> > > foo
bar`, `
<blockquote>
<blockquote>
<blockquote>
<p>foo
bar</p>
</blockquote>
</blockquote>
</blockquote>`},
	{"198", `
>>> foo
> bar
>>baz`, `
<blockquote>
<blockquote>
<blockquote>
<p>foo
bar
baz</p>
</blockquote>
</blockquote>
</blockquote>`},
	{"199", `
>     code

>    not code`, `
<blockquote>
<pre><code>code
</code></pre>
</blockquote>
<blockquote>
<p>not code</p>
</blockquote>`},
	{"200", `
A paragraph
with two lines.

    indented code

> A block quote.`, `
<p>A paragraph
with two lines.</p>
<pre><code>indented code
</code></pre>
<blockquote>
<p>A block quote.</p>
</blockquote>`},
	{"201", `
1.  A paragraph
    with two lines.

        indented code

    > A block quote.`, `
<ol>
<li>
<p>A paragraph
with two lines.</p>
<pre><code>indented code
</code></pre>
<blockquote>
<p>A block quote.</p>
</blockquote>
</li>
</ol>`},
	{"203", `
- one

  two`, `
<ul>
<li>
<p>one</p>
<p>two</p>
</li>
</ul>`},
	{"205", `
 -    one

      two`, `
<ul>
<li>
<p>one</p>
<p>two</p>
</li>
</ul>`},
	{"206", `
   > > 1.  one
>>
>>     two`, `
<blockquote>
<blockquote>
<ol>
<li>
<p>one</p>
<p>two</p>
</li>
</ol>
</blockquote>
</blockquote>`},
	{"207", `
>>- one
>>
  >  > two`, `
<blockquote>
<blockquote>
<ul>
<li>one</li>
</ul>
<p>two</p>
</blockquote>
</blockquote>`},
	{"208", `-one

2.two`, `
<p>-one</p>
<p>2.two</p>`},
	{"210", `
1.  foo

    ~~~
    bar
    ~~~

    baz

    > bam`, `
<ol>
<li>
<p>foo</p>
<pre><code>bar
</code></pre>
<p>baz</p>
<blockquote>
<p>bam</p>
</blockquote>
</li>
</ol>`},
	{"212", `1234567890. not ok`, `<p>1234567890. not ok</p>`},
	{"215", `-1. not ok`, `<p>-1. not ok</p>`},
	{"216", `
- foo

      bar`, `
<ul>
<li>
<p>foo</p>
<pre><code>bar
</code></pre>
</li>
</ul>`},
	{"218", `
    indented code

paragraph

    more code`, `
<pre><code>indented code
</code></pre>
<p>paragraph</p>
<pre><code>more code
</code></pre>`},
	{"221", `
   foo

bar`, `
<p>foo</p>
<p>bar</p>`},
	{"223", `
-  foo

   bar`, `
<ul>
<li>
<p>foo</p>
<p>bar</p>
</li>
</ul>`},
	{"226", `
- foo
-   
- bar`, `
<ul>
<li>foo</li>
<li></li>
<li>bar</li>
</ul>`},
	{"232", `
    1.  A paragraph
        with two lines.

            indented code

        > A block quote.`, `
<pre><code>1.  A paragraph
    with two lines.

        indented code

    &gt; A block quote.
</code></pre>`},
	{"234", `
  1.  A paragraph
    with two lines.`, `
<ol>
<li>A paragraph
with two lines.</li>
</ol>`},
	{"235", `
> 1. > Blockquote
continued here.`, `
<blockquote>
<ol>
<li>
<blockquote>
<p>Blockquote
continued here.</p>
</blockquote>
</li>
</ol>
</blockquote>`},
	{"236", `
> 1. > Blockquote
continued here.`, `
<blockquote>
<ol>
<li>
<blockquote>
<p>Blockquote
continued here.</p>
</blockquote>
</li>
</ol>
</blockquote>`},
	{"237", `
- foo
  - bar
    - baz`, `
<ul>
<li>foo
<ul>
<li>bar
<ul>
<li>baz</li>
</ul>
</li>
</ul>
</li>
</ul>`},
	{"241", "- - foo", `
<ul>
<li>
<ul>
<li>foo</li>
</ul>
</li>
</ul>`},
	{"243", `
- # Foo
- Bar
  ---
  baz`, `
<ul>
<li>
<h1>Foo</h1>
</li>
<li>
<h2>Bar</h2>
baz</li>
</ul>`},
	{"246", `
Foo
- bar
- baz`, `
<p>Foo</p>
<ul>
<li>bar</li>
<li>baz</li>
</ul>`},
	{"248", `
- foo

- bar


- baz`, `
<ul>
<li>
<p>foo</p>
</li>
<li>
<p>bar</p>
</li>
</ul>
<ul>
<li>baz</li>
</ul>`},
	{"249", `
- foo


  bar
- baz`, `
<ul>
<li>foo</li>
</ul>
<p>bar</p>
<ul>
<li>baz</li>
</ul>`},
	{"250", `
- foo
  - bar
    - baz


      bim`, `
<ul>
<li>foo
<ul>
<li>bar
<ul>
<li>baz</li>
</ul>
</li>
</ul>
</li>
</ul>
<pre><code>  bim
</code></pre>`},
	{"251", `
- foo
- bar


- baz
- bim`, `
<ul>
<li>foo</li>
<li>bar</li>
</ul>
<ul>
<li>baz</li>
<li>bim</li>
</ul>`},
	{"252", `
-   foo

    notcode

-   foo


    code`, `
<ul>
<li>
<p>foo</p>
<p>notcode</p>
</li>
<li>
<p>foo</p>
</li>
</ul>
<pre><code>code
</code></pre>`},
	{"261", `
* a
  > b
  >
* c`, `
<ul>
<li>a
<blockquote>
<p>b</p>
</blockquote>
</li>
<li>c</li>
</ul>`},
	{"263", "- a", `
<ul>
<li>a</li>
</ul>`},
	{"264", `
- a
  - b`, `
<ul>
<li>a
<ul>
<li>b</li>
</ul>
</li>
</ul>`},
	{"265", "\n1. ```\n   foo\n   ```\n\n   bar", `
<ol>
<li>
<pre><code>foo
</code></pre>
<p>bar</p>
</li>
</ol>`},
	{"266", `
* foo
  * bar

  baz`, `
<ul>
<li>
<p>foo</p>
<ul>
<li>bar</li>
</ul>
<p>baz</p>
</li>
</ul>`},
	{"267", `
- a
  - b
  - c

- d
  - e
  - f`, `
<ul>
<li>
<p>a</p>
<ul>
<li>b</li>
<li>c</li>
</ul>
</li>
<li>
<p>d</p>
<ul>
<li>e</li>
<li>f</li>
</ul>
</li>
</ul>`},
	{"268", "`hi`lo`", "<p><code>hi</code>lo`</p>"},
	{"273", `
foo\
bar
`, `
<p>foo<br>
bar</p>`},
	{"275", `    \[\]`, `<pre><code>\[\]
</code></pre>`},
	{"276", `
~~~
\[\]
~~~`, `
<pre><code>\[\]
</code></pre>`},
	{"294", "`foo`", `<p><code>foo</code></p>`},
	{"300", "`foo\\`bar`", "<p><code>foo\\</code>bar`</p>"},
	{"303", "`<a href=\"`\">`", "<p><code>&lt;a href=&quot;</code>&quot;&gt;`</p>"},
	{"308", "`foo", "<p>`foo</p>"},
	{"309", "*foo bar*", "<p><em>foo bar</em></p>"},
	{"310", "a * foo bar*", "<p>a * foo bar*</p>"},
	{"313", "foo*bar*", "<p>foo<em>bar</em></p>"},
	{"314", "5*6*78", "<p>5<em>6</em>78</p>"},
	{"315", "_foo bar_", "<p><em>foo bar</em></p>"},
	{"316", "_ foo bar_", "<p>_ foo bar_</p>"},
	{"322", "foo-_(bar)_", "<p>foo-<em>(bar)</em></p>"},
	{"323", "_foo*", "<p>_foo*</p>"},
	{"328", "*foo*bar", "<p><em>foo</em>bar</p>"},
	{"335", "_(bar)_.", "<p><em>(bar)</em>.</p>"},
	{"336", "**foo bar**", "<p><strong>foo bar</strong></p>"},
	{"339", "foo**bar**", "<p>foo<strong>bar</strong></p>"},
	{"340", "__foo bar__", "<p><strong>foo bar</strong></p>"},
	{"348", "foo-__(bar)__", "<p>foo-<strong>(bar)</strong></p>"},
	{"352", "**Gomphocarpus (*Gomphocarpus physocarpus*, syn.*Asclepias physocarpa*)**",
		"<p><strong>Gomphocarpus (<em>Gomphocarpus physocarpus</em>, syn.<em>Asclepias physocarpa</em>)</strong></p>"},
	{"353", "**foo \"*bar*\" foo**", "<p><strong>foo &quot;<em>bar</em>&quot; foo</strong></p>"},
	{"354", "**foo**bar", "<p><strong>foo</strong>bar</p>"},
	{"361", "__(bar)__.", "<p><strong>(bar)</strong>.</p>"},
	{"362", "*foo [bar](/url)*", "<p><em>foo <a href=\"/url\">bar</a></em></p>"},
	{"363", "*foo\nbar*", "<p><em>foo\nbar</em></p>"},
	{"375", "** is not an empty emphasis", "<p>** is not an empty emphasis</p>"},
	{"377", "**foo [bar](/url)**", "<p><strong>foo <a href=\"/url\">bar</a></strong></p>"},
	{"378", "**foo\nbar**", "<p><strong>foo\nbar</strong></p>"},
	{"379", "__foo _bar_ baz__", "<p><strong>foo <em>bar</em> baz</strong></p>"},
	{"383", "**foo *bar* baz**", "<p><strong>foo <em>bar</em> baz</strong></p>"},
	{"385", "***foo* bar**", "<p><strong><em>foo</em> bar</strong></p>"},
	{"386", "**foo *bar***", "<p><strong>foo <em>bar</em></strong></p>"},
	{"389", "__ is not an empty emphasis", "<p>__ is not an empty emphasis</p>"},
	{"392", "foo *\\**", "<p>foo <em>*</em></p>"},
	{"393", "foo *_*", "<p>foo <em>_</em></p>"},
	{"395", "foo **\\***", "<p>foo <strong>*</strong></p>"},
	{"396", "foo **_**", "<p>foo <strong>_</strong></p>"},
	{"404", "foo _\\__", "<p>foo <em>_</em></p>"},
	{"405", "foo _*_", "<p>foo <em>*</em></p>"},
	{"407", "foo __\\___", "<p>foo <strong>_</strong></p>"},
	{"408", "foo __*__", "<p>foo <strong>*</strong></p>"},
	{"415", "**foo**", "<p><strong>foo</strong></p>"},
	{"416", "*_foo_*", "<p><em><em>foo</em></em></p>"},
	{"417", "__foo__", "<p><strong>foo</strong></p>"},
	{"418", "_*foo*_", "<p><em><em>foo</em></em></p>"},
	{"419", "****foo****", "<p><strong><strong>foo</strong></strong></p>"},
	{"420", "____foo____", "<p><strong><strong>foo</strong></strong></p>"},
	{"422", "***foo***", "<p><strong><em>foo</em></strong></p>"},
	{"424", "*foo _bar* baz_", "<p><em>foo _bar</em> baz_</p>"},
	{"438", "[link](/uri \"title\")", "<p><a href=\"/uri\" title=\"title\">link</a></p>"},
	{"439", "[link](/uri)", "<p><a href=\"/uri\">link</a></p>"},
	{"440", "[link]()", "<p><a href=\"\">link</a></p>"},
	{"441", "[link](<>)", "<p><a href=\"\">link</a></p>"},
	{"451", `
[link](#fragment)

[link](http://example.com#fragment)

[link](http://example.com?foo=bar&baz#fragment)`, `
<p><a href="#fragment">link</a></p>
<p><a href="http://example.com#fragment">link</a></p>
<p><a href="http://example.com?foo=bar&amp;baz#fragment">link</a></p>`},
	{"455", `
[link](/url "title")
[link](/url 'title')
[link](/url (title))`, `
<p><a href="/url" title="title">link</a>
<a href="/url" title="title">link</a>
<a href="/url" title="title">link</a></p>`},
	{"458", `[link](/url 'title "and" title')`, `<p><a href="/url" title="title &quot;and&quot; title">link</a></p>`},
	{"460", "[link] (/uri)", "<p>[link] (/uri)</p>"},
	{"461", "[link [foo [bar]]](/uri)", `<p><a href="/uri">link [foo [bar]]</a></p>`},
	{"463", "[link [bar](/uri)", `<p>[link <a href="/uri">bar</a></p>`},
	{"471", "[foo *bar](baz*)", `<p><a href="baz*">foo *bar</a></p>`},
	{"472", "*foo [bar* baz]", "<p><em>foo [bar</em> baz]</p>"},
	{"476", `
[foo][bar]

[bar]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"477", `
[link [foo [bar]]][ref]

[ref]: /uri`, `<p><a href="/uri">link [foo [bar]]</a></p>`},
	{"484", `
[foo *bar][ref]

[ref]: /uri`, `<p><a href="/uri">foo *bar</a></p>`},
	{"488", `
[foo][BaR]

[bar]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"489", `
[Толпой][Толпой] is a Russian word.

[ТОЛПОЙ]: /url`, `<p><a href="/url">Толпой</a> is a Russian word.</p>`},
	{"491", `
[foo] [bar]

[bar]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"492", `
[foo]
[bar]

[bar]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"493", `
[foo]: /url1

[foo]: /url2

[bar][foo]`, `<p><a href="/url1">bar</a></p>`},
	{"496", `
[foo][ref[bar]]

[ref[bar]]: /uri`, `
<p>[foo][ref[bar]]</p>
<p>[ref[bar]]: /uri</p>`},
	{"497", `
[[[foo]]]

[[[foo]]]: /url`, `
<p>[[[foo]]]</p>
<p>[[[foo]]]: /url</p>`},
	{"498", `
[foo][ref\[]

[ref\[]: /uri`, `<p><a href="/uri">foo</a></p>`},
	{"499", `
[]

[]: /uri`, `
<p>[]</p>
<p>[]: /uri</p>`},
	{"501", `
[foo][]

[foo]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"502", `
[*foo* bar][]

[*foo* bar]: /url "title"`, `
<p><a href="/url" title="title"><em>foo</em> bar</a></p>`},
	{"503", `
[Foo][]

[foo]: /url "title"`, `<p><a href="/url" title="title">Foo</a></p>`},
	{"504", `
[foo] 
[]

[foo]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"505", `
[foo]

[foo]: /url "title"`, `<p><a href="/url" title="title">foo</a></p>`},
	{"506", `
[*foo* bar]

[*foo* bar]: /url "title"`, `
<p><a href="/url" title="title"><em>foo</em> bar</a></p>`},
	{"508", `
[[bar [foo]

[foo]: /url`, `<p>[[bar <a href="/url">foo</a></p>`},
	{"509", `
[Foo]

[foo]: /url "title"`, `<p><a href="/url" title="title">Foo</a></p>`},
	{"510", `
[foo] bar

[foo]: /url`, `<p><a href="/url">foo</a> bar</p>`},
	{"511", `
\[foo]

[foo]: /url "title"`, `<p>[foo]</p>`},
	{"513", `
[foo][bar]

[foo]: /url1
[bar]: /url2`, `<p><a href="/url2">foo</a></p>`},
	{"515", `
[foo][bar][baz]

[baz]: /url1
[bar]: /url2`, `<p><a href="/url2">foo</a><a href="/url1">baz</a></p>`},
	{"517", `![foo](/url "title")`, `<p><img src="/url" alt="foo" title="title"></p>`},
	{"523", `![foo](train.jpg)`, `<p><img src="train.jpg" alt="foo"></p>`},
	{"524", `My ![foo bar](/path/to/train.jpg  "title"   )`,
		`<p>My <img src="/path/to/train.jpg" alt="foo bar" title="title"></p>`},
	{"525", `![foo](<url>)`, `<p><img src="url" alt="foo"></p>`},
	{"526", `![](/url)`, `<p><img src="/url" alt=""></p>`},
	{"527", `
![foo] [bar]

[bar]: /url`, `<p><img src="/url" alt="foo"></p>`},
	{"528", `
![foo] [bar]

[BAR]: /url`, `<p><img src="/url" alt="foo"></p>`},
	{"529", `
![foo][]

[foo]: /url "title"`, `<p><img src="/url" alt="foo" title="title"></p>`},
	{"531", `
![Foo][]

[foo]: /url "title"`, `<p><img src="/url" alt="Foo" title="title"></p>`},
	{"532", `
![foo] 
[]

[foo]: /url "title"`, `<p><img src="/url" alt="foo" title="title"></p>`},
	{"533", `
![foo]

[foo]: /url "title"`, `<p><img src="/url" alt="foo" title="title"></p>`},
	{"535", `
![[foo]]

[[foo]]: /url "title"`, `
<p>![[foo]]</p>
<p>[[foo]]: /url &quot;title&quot;</p>`},
	{"536", `
![Foo]

[foo]: /url "title"`, `<p><img src="/url" alt="Foo" title="title"></p>`},
	{"537", `
\!\[foo]

[foo]: /url "title"`, `<p>![foo]</p>`},
	{"538", `
\![foo]

[foo]: /url "title"`, `<p>!<a href="/url" title="title">foo</a></p>`},
	{"539", `<http://foo.bar.baz>`, `<p><a href="http://foo.bar.baz">http://foo.bar.baz</a></p>`},
	{"540", `<http://foo.bar.baz/test?q=hello&id=22&boolean>`,
		`<p><a href="http://foo.bar.baz/test?q=hello&amp;id=22&amp;boolean">http://foo.bar.baz/test?q=hello&amp;id=22&amp;boolean</a></p>`},
	{"541", `<irc://foo.bar:2233/baz>`, `<p><a href="irc://foo.bar:2233/baz">irc://foo.bar:2233/baz</a></p>`},
	{"542", `<MAILTO:FOO@BAR.BAZ>`, `<p><a href="MAILTO:FOO@BAR.BAZ">MAILTO:FOO@BAR.BAZ</a></p>`},
	{"548", "<>", "<p>&lt;&gt;</p>"},
	{"554", `foo@bar.example.com`, `<p>foo@bar.example.com</p>`},
	{"555", "<a><bab><c2c>", "<p><a><bab><c2c></p>"},
	{"556", "<a/><b2/>", "<p><a/><b2/></p>"},
	{"557", `
<a  /><b2
data="foo" >`, `
<p><a  /><b2
data="foo" ></p>`},
	{"558", `
<a foo="bar" bam = 'baz <em>"</em>'
_boolean zoop:33=zoop:33 />`, `
<p><a foo="bar" bam = 'baz <em>"</em>'
_boolean zoop:33=zoop:33 /></p>`},
	{"572", "foo <![CDATA[>&<]]>", "<p>foo <![CDATA[>&<]]></p>"},
	{"576", `
foo  
baz`, `
<p>foo<br>
baz</p>`},
	{"577", `
foo\
baz`, `
<p>foo<br>
baz</p>`},
	{"578", `
foo       
baz`, `<p>foo<br>baz</p>`},
	{"581", `
*foo  
bar*`, `
<p><em>foo<br>
bar</em></p>`},
	{"582", `
*foo\
bar*`, `
<p><em>foo<br>
bar</em></p>`},
	{"587", `foo\`, `<p>foo\</p>`},
	{"588", `foo  `, `<p>foo</p>`},
	{"589", `### foo\`, `<h3>foo\</h3>`},
	{"590", `### foo  `, `<h3>foo</h3>`},
	{"591", `
foo
baz`, `
<p>foo
baz</p>`},
	{"592", `
foo 
 baz`, `
<p>foo
baz</p>`},
	{"594", `Foo χρῆν`, `<p>Foo χρῆν</p>`},
	{"595", `Multiple     spaces`, `<p>Multiple     spaces</p>`},
}

func TestCommonMark(t *testing.T) {
	reID := regexp.MustCompile(` +?id=".*"`)
	for _, c := range CMCases {
		// Remove the auto-hashing until it'll be in the configuration
		actual := reID.ReplaceAllString(Render(c.input), "")
		if strings.Replace(actual, "\n", "", -1) != strings.Replace(c.expected, "\n", "", -1) {
			t.Errorf("\ninput:%s\ngot:\n%s\nexpected:\n%s\nlink: http://spec.commonmark.org/0.21/#example-%s\n",
				c.input, actual, c.expected, c.name)
		}
	}
}
