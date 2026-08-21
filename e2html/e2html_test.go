package e2html

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAAndAttrSet(t *testing.T) {
	a := A("id", "main")
	assert.Equal(t, "main", a["id"])

	a = a.Set("class", "box")
	assert.Equal(t, "box", a["class"])

	var nilAttr Attr
	nilAttr = nilAttr.Set("k", "v")
	require.NotNil(t, nilAttr)
	assert.Equal(t, "v", nilAttr["k"])
}

func TestAttrString(t *testing.T) {
	assert.Equal(t, "", Attr{}.String())
	assert.Equal(t, ` class="title" id="n1"`, Attr{"id": "n1", "class": "title"}.String())
	assert.Equal(t, ` disabled="disabled"`, Attr{"disabled": true}.String())
	assert.Equal(t, "", Attr{"hidden": false}.String())
	assert.Equal(t, ` disabled`, Attr{"disabled": nil}.String())
	assert.Contains(t, Attr{"title": `<script>`}.String(), `title="&lt;script&gt;"`)
}

func TestTBasic(t *testing.T) {
	got := T("div", Attr{"id": "n1"}, Text("hello")).String()
	assert.Equal(t, `<div id="n1">hello</div>`, got)

	got = T("p", "plain").String()
	assert.Equal(t, `<p>plain</p>`, got)

	got = T("span", 42).String()
	assert.Equal(t, `<span>42</span>`, got)
}

func TestTNestedAndSlice(t *testing.T) {
	inner := []TAG{
		T("li", Text("a")),
		T("li", Text("b")),
	}
	got := T("ul", inner).String()
	assert.Equal(t, `<ul><li>a</li><li>b</li></ul>`, got)

	got = T("div", T("span", "x"), T("span", "y")).String()
	assert.Equal(t, `<div><span>x</span><span>y</span></div>`, got)
}

func TestTEscapesText(t *testing.T) {
	got := T("div", Text(`<script>alert("x")</script>`)).String()
	assert.NotContains(t, got, `<script>alert`)
	assert.Contains(t, got, `&lt;script&gt;`)
}

func TestTLastTextWins(t *testing.T) {
	got := T("p", "first", "second").String()
	assert.Equal(t, `<p>second</p>`, got)
}

func TestVoidElements(t *testing.T) {
	assert.Equal(t, `<br>`, T("br").String())
	assert.Equal(t, `<img src="/a.png">`, T("img", A("src", "/a.png")).String())
	assert.Equal(t, `<input type="text">`, T("input", Attr{"type": "text"}, Text("ignored")).String())
	assert.Equal(t, `<HR>`, T("HR").String())
	assert.False(t, strings.Contains(T("img", A("src", "x")).String(), "</img>"))
}

func TestInvalidTagName(t *testing.T) {
	got := T(`div onclick=alert(1)`, Text("x")).String()
	assert.False(t, strings.HasPrefix(got, "<"))
	assert.NotContains(t, got, "<div")
	assert.Contains(t, got, "onclick=alert(1)")

	got = T("ok-widget", Text("hi")).String()
	assert.Equal(t, `<ok-widget>hi</ok-widget>`, got)
}

func TestComment(t *testing.T) {
	got := T("<!--", Text("note")).String()
	assert.Contains(t, got, "<!--")
	assert.Contains(t, got, "note")
	assert.Contains(t, got, "-->")
}

func TestTS(t *testing.T) {
	assert.Equal(t, TAG("<p>a</p>"), TS(T("p", "a")))
	assert.Equal(t, TAG("<p>a</p><p>b</p>"), TS([]TAG{T("p", "a"), T("p", "b")}))
}

func TestDivAndDoctype(t *testing.T) {
	assert.Equal(t, `<div id="x">hi</div>`, Div(A("id", "x"), Text("hi")).String())
	assert.Equal(t, `<!DOCTYPE html>`, Doctype("html"))
	assert.Equal(t, `<!DOCTYPE html>`, Doctype(`html><script>`))
	assert.Equal(t, `<!DOCTYPE HTML PUBLIC>`, Doctype("HTML PUBLIC"))
}

func TestSelectOptions(t *testing.T) {
	options := []TAG{
		T("option", Attr{"value": "1", "selected": false}, Text("one")),
		T("option", Attr{"value": "2", "selected": true}, Text("two")),
	}
	got := T("select", options).String()
	assert.Equal(t, `<select><option value="1">one</option><option selected="selected" value="2">two</option></select>`, got)
}
