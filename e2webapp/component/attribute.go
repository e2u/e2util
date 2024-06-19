package component

import (
	"fmt"
	"strings"
)

type Attributes map[string]any

func (attr Attributes) String() string {
	var buf []string
	for k, v := range attr {
		ak := strings.ToLower(k)
		if ak == "selected" || ak == "checked" {
			buf = append(buf, fmt.Sprintf(`%s`, k))
		} else {
			buf = append(buf, fmt.Sprintf(`%s="%v"`, k, v))
		}
	}
	return " " + strings.Join(buf, " ")
}
