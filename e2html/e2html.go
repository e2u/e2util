package e2html

import (
	"fmt"
	"html"
	"html/template"
	"maps"
	"sort"
	"strings"
)

type A map[string]any

func (attr A) String() string {
	var buf []string
	var keys []string
	for k := range attr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := attr[k]
		switch v := v.(type) {
		case bool:
			if v {
				buf = append(buf, fmt.Sprintf(`%s="%s"`, k, k))
			}
		case nil:
			buf = append(buf, k)
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

func (r TAG) HTML() template.HTML {
	return template.HTML(r) // #nosec G203
}

func TS[T TAG | []TAG](t T) TAG {
	var rs []string
	if v, ok := any(t).(TAG); ok {
		return v
	}

	for _, v := range any(t).([]TAG) {
		rs = append(rs, v.String())
	}
	return TAG(strings.Join(rs, ""))
}

func T(name string, args ...any) TAG {
	name = strings.TrimSpace(name)
	isComment := strings.HasPrefix(name, "<!--")
	name = escape(name)

	var rs []string
	attrs := make(A)
	var text Text
	var subResult []TAG

	for _, arg := range args {
		switch v := arg.(type) {
		case TAG:
			subResult = append(subResult, v)
		case []TAG:
			subResult = append(subResult, v...)
		case Text,
			string:
			text = Text(escape(v))
		case A:
			maps.Copy(attrs, v)
		default:
			text = Text(escape(v))
		}
	}

	for _, sub := range subResult {
		rs = append(rs, sub.String())
	}

	if !isComment {
		rs = append([]string{fmt.Sprintf(`<%s%s>%s`, name, attrs.String(), text)}, rs...)
		rs = append(rs, fmt.Sprintf(`</%s>`, name))
		return TAG(strings.Join(rs, ""))
	}

	rs = append([]string{fmt.Sprintf("\n\n\n<!--\n\n%s", text)}, rs...)
	rs = append(rs, "\n\n-->\n\n\n")
	return TAG(strings.Join(rs, ""))
}

func escape(s any) string {
	return html.EscapeString(fmt.Sprintf("%v", s))
}
