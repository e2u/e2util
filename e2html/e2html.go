package e2html

import (
	"fmt"
	"html"
	"html/template"
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
		switch v := v.(type) {
		case bool:
			if v {
				buf = append(buf, fmt.Sprintf(`%s="%s"`, k, k))
			}
		case nil:
			buf = append(buf, escape(k))
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

// dangerousPatterns contains regex patterns that indicate potentially dangerous HTML content
var dangerousPatterns = []string{
	`(?i)<script[^>]*>`,           // script tags
	`(?i)</script>`,                // closing script tags
	`(?i)javascript:`,              // javascript protocol
	`(?i)on\w+\s*=`,                // event handlers (onclick, onload, etc.)
	`(?i)<iframe[^>]*>`,            // iframes
	`(?i)<object[^>]*>`,            // objects
	`(?i)<embed[^>]*>`,             // embeds
}

// containsDangerousContent checks if the string contains potentially dangerous HTML
func containsDangerousContent(s string) bool {
	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, s); matched {
			return true
		}
	}
	return false
}

// HTML returns the TAG as template.HTML.
// WARNING: This method validates the content for dangerous patterns (scripts, event handlers, etc.)
// but should still be used with caution. Only use with trusted or sanitized content.
// For untrusted content, use String() instead which properly escapes the output.
// Returns empty string if dangerous content is detected.
func (r TAG) HTML() template.HTML {
	if containsDangerousContent(string(r)) {
		return ""
	}
	return template.HTML(r) // #nosec G203
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

	// Should never reach here due to type constraint, but handle gracefully
	return TAG("")
}

func T(name string, args ...any) TAG {
	name = strings.TrimSpace(name)
	isComment := strings.HasPrefix(name, "<!--")
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
		case Text,
			string:
			text = Text(escape(v))
		case Attr:
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
