package e2regexp

import (
	"regexp"
)

func NamedFindStringSubmatch(s string, r *regexp.Regexp) (map[string]string, bool) {
	rs := make(map[string]string)
	if !r.MatchString(s) {
		return rs, false
	}
	match := r.FindStringSubmatch(s)
	for i, name := range r.SubexpNames() {
		rs[name] = match[i]
	}
	return rs, true
}
