package e2html

import (
	"fmt"
	"html"
	"maps"
	"regexp"
	"sort"
	"strings"
)

type Attr map[string]any

func A(key string, value any) Attr {
	ra := make(Attr)
	ra.Set(key, value)
	return ra
}

func (attr Attr) Set(key string, value any) Attr {
	if attr == nil {
		attr = make(Attr)
	}
	attr[key] = value
	return attr
}

func (attr Attr) String() string {
	var buf []string
	var keys []string
	for k := range attr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := attr[k]
		ek := escape(k)
		switch v := v.(type) {
		case bool:
			if v {
				buf = append(buf, fmt.Sprintf(`%s="%s"`, ek, ek))
			}
		case nil:
			buf = append(buf, ek)
		default:
			buf = append(buf, fmt.Sprintf(`%s="%s"`, ek, escape(v)))
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

// TS converts a single TAG or a slice of TAG to a single TAG.
// Returns an empty TAG if the input type is unexpected.
func TS[T TAG | []TAG](t T) TAG {
	var rs []string
	if v, ok := any(t).(TAG); ok {
		return v
	}

	if v, ok := any(t).([]TAG); ok {
		for _, tag := range v {
			rs = append(rs, tag.String())
		}
		return TAG(strings.Join(rs, ""))
	}

	return TAG("")
}

var validTagName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9:-]*$`)

var voidElements = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "embed": {},
	"hr": {}, "img": {}, "input": {}, "link": {}, "meta": {},
	"param": {}, "source": {}, "track": {}, "wbr": {},
}

func isVoidElement(name string) bool {
	_, ok := voidElements[strings.ToLower(name)]
	return ok
}

func T(name string, args ...any) TAG {
	name = strings.TrimSpace(name)
	isComment := strings.HasPrefix(name, "<!--")
	if !isComment && !validTagName.MatchString(name) {
		return TAG(escape(name))
	}

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
		case Text, string:
			text = Text(escape(v))
		case Attr:
			maps.Copy(attrs, v)
		default:
			text = Text(escape(v))
		}
	}

	if isComment {
		rs = append([]string{fmt.Sprintf("\n\n\n<!--\n\n%s", text)}, rs...)
		rs = append(rs, "\n\n-->\n\n\n")
		return TAG(strings.Join(rs, ""))
	}

	if isVoidElement(name) {
		return TAG(fmt.Sprintf(`<%s%s>`, name, attrs.String()))
	}

	for _, sub := range subResult {
		rs = append(rs, sub.String())
	}
	rs = append([]string{fmt.Sprintf(`<%s%s>%s`, name, attrs.String(), text)}, rs...)
	rs = append(rs, fmt.Sprintf(`</%s>`, name))
	return TAG(strings.Join(rs, ""))
}

func escape(s any) string {
	return html.EscapeString(fmt.Sprintf("%v", s))
}
