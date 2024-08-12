package e2test

import (
	"fmt"
	"strings"
)

func Line(params ...any) {
	var s string
	var n int
	var skip bool
	for _, param := range params {
		switch v := param.(type) {
		case string:
			s = v
		case int:
			n = v
		case bool:
			skip = v
		}
	}
	if n == 0 {
		n = 80
	}
	if !skip {
		fmt.Println(strings.Repeat(s, n))
	}
}
