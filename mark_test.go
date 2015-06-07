package mark

import (
	"github.com/a8m/expect"
	"testing"
)

func TestRender(t *testing.T) {
	expect := expect.New(t)
	cases := map[string]string{
		"foobar":                "<p>foobar</p>",
		"foo|bar":               "<p>foo|bar</p>",
		"foo  \nbar":            "<p>foo<br>bar</p>",
		"__bar__ foo":           "<p><strong>bar</strong> foo</p>",
		"**bar** foo __bar__":   "<p><strong>bar</strong> foo <strong>bar</strong></p>",
		"**bar**__baz__":        "<p><strong>bar</strong><strong>baz</strong></p>",
		"**bar**foo__bar__":     "<p><strong>bar</strong>foo<strong>bar</strong></p>",
		"_bar_baz":              "<p><em>bar</em>baz</p>",
		"_foo_~~bar~~ baz":      "<p><em>foo</em><del>bar</del> baz</p>",
		"~~baz~~ _baz_":         "<p><del>baz</del> <em>baz</em></p>",
		"`bool` and that's it.": "<p><code>bool</code> and that's it.</p>",
		//	"___mixim___": "<p><strong><em>foo</em></strong></p>",
		// Paragraph
		"1  \n2  \n3":        "<p>1<br>2<br>3</p>",
		"1\n\n2":             "<p>1</p>\n<p>2</p>",
		"1\n\n\n2":           "<p>1</p>\n<p>2</p>",
		"1\n\n\n\n\n\n\n\n2": "<p>1</p>\n<p>2</p>",
		// Heading
		"#1\n##2":                  "<h1>1</h1>\n<h2>2</h2>",
		"#1\np\n##2\n###3\n4\n===": "<h1>1</h1>\n<p>p</p>\n<h2>2</h2>\n<h3>3</h3>\n<h1>4</h1>",
		"Hello\n===":               "<h1>Hello</h1>",
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
		"Link: <not really":      "<p>Link: <not really</p>",
		// CodeBlock
		"\tfoo\n\tbar": "<pre><code>foo\nbar</code></pre>",
		"\tfoo\nbar":   "<pre><code>foo\n</code></pre><p>bar</p>",
		// GfmCodeBlock
		"```js\nvar a;\n```":  "<pre><code class=\"lang-js\">var a;</code></pre>",
		"~~~\nvar b;~~~":      "<pre><code>var b;</code></pre>",
		"~~~js\nlet d = 1~~~": "<pre><code>let d = 1</code></pre>",
		// Hr
		"foo\n****\nbar": "<p>foo</p>\n<hr><p>bar</p>",
		"foo\n___":       "<p>foo</p>\n<hr>",
		// Images
		"![name](url)":           "<p><img src=\"url\" alt=\"name\"></p>",
		"![name](url \"title\")": "<p><img src=\"url\" alt=\"name\" title=\"title\"></p>",
		"img: ![name]()":         "<p>img: <img src=\"\" alt=\"name\"></p>",
		// Lists
		"- foo\n- bar": "<ul><li>foo</li><li>bar</li></ul>",
		"* foo\n* bar": "<ul><li>foo</li><li>bar</li></ul>",
		"+ foo\n+ bar": "<ul><li>foo</li><li>bar</li></ul>",
		// Ordered Lists
		"1. one\n2. two\n3. three": "<ol><li>one</li><li>two</li><li>three</li></ol>",
		"1. one\n 1. one of one":   "<ol><li>one<ol><li>one of one</li></ol></li></ol>",
		"2. two\n 3. three":        "<ol><li>two<ol><li>three</li></ol></li></ol>",
	}
	for actual, expected := range cases {
		expect(Render(actual)).To.Equal(expected)
	}
}
