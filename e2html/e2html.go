package e2html

import (
	"fmt"
	"html"
	"maps"
	"strings"
)

type Attr map[string]any

func (attr Attr) String() string {
	var buf []string
	for k, v := range attr {
		switch v := v.(type) {
		case bool:
			if v {
				buf = append(buf, fmt.Sprintf(`%s="%s"`, k, k))
			}
		case nil:
			buf = append(buf, fmt.Sprintf(`%s`, k))
		default:
			buf = append(buf, fmt.Sprintf(`%s="%v"`, k, escape(v)))
		}
	}
	if len(buf) == 0 {
		return ""
	}
	return " " + strings.Join(buf, " ")
}

type Text string

func (t Text) String() string {
	return string(t)
}

type TAG string

func (r TAG) String() string {
	return string(r)
}

func Tag(name string, args ...any) TAG {
	name = escape(name)
	var rs []string
	attrs := make(Attr)
	var text Text
	var subResult []TAG
	for _, arg := range args {
		switch v := arg.(type) {
		case TAG:
			subResult = append(subResult, v)
		case []TAG:
			subResult = append(subResult, v...)
		case Text:
			text = Text(escape(v))
		case Attr:
			maps.Copy(attrs, v)
		}
	}
	rs = append(rs, fmt.Sprintf(`<%s%s>%s`, name, attrs.String(), text))
	for _, sub := range subResult {
		rs = append(rs, sub.String())
	}
	rs = append(rs, fmt.Sprintf(`</%s>`, name))
	return TAG(strings.Join(rs, ""))
}

func escape(s any) string {
	return html.EscapeString(fmt.Sprintf("%v", s))
}
