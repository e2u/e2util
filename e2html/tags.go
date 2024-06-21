// https://www.w3schools.com/tags/tag_comment.asp

package e2html

import (
	"fmt"
)

func Doctype(t string) string {
	return fmt.Sprintf("<!DOCTYPE %s>", t)
}

func Div(args ...any) TAG {
	return T("div", args...)
}

// ....
