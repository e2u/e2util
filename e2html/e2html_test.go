package e2html

import (
	"fmt"
	"testing"
)

func Test_Tag(t *testing.T) {
	subDiv1 := Tag("div", Attr{"id": "sub-div-01", "class": "sub-title", "disabled": nil}, Text(">>>Su<<<bDivText"))
	subDiv2 := Tag("div", Attr{"id": "sub-div-02", "class": "sub-title"}, Text("SubDivText"))
	subDiv3 := Tag("div", Attr{"id": "sub-div-03", "class": "sub-title"}, Text("SubDivText"))
	div := Tag("div", Attr{"id": "div-01", "class": "title"}, subDiv1, subDiv2, subDiv3)

	fmt.Println(div)
}

func Test_Div(t *testing.T) {
	sd1 := Div(Attr{"id": "sub-div-01", "class": "sub-title", "disabled": nil}, Text(">>>Su<<<bDivText"))
	sd2 := Div(Attr{"id": "sub-div-02", "class": "sub-title"}, Text("SubDivText"))
	d := Div(Attr{"id": "div-01", "class": "title"}, sd1, sd2)
	fmt.Println(d)
}

func Test_Select(t *testing.T) {
	options := []TAG{
		Tag("option", Attr{"value": "value-01", "selected": false}, Text("option value 1")),
		Tag("option", Attr{"value": "value-02", "selected": true}, Text("option value 2")),
		Tag("option", Attr{"value": "value-03", "selected": false}, Text("option value 3")),
	}
	fmt.Println(Tag("select", options))
}
