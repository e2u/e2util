package e2regexp

import (
	"regexp"

	"github.com/e2u/e2util/e2map"
)

/*
*
phone := "202-555-0147"
phoneRENamedCaps := `(?P<area>\d{3})\-(?P<exchange>\d{3})\-(?P<line>\d{4})$`
re = regexp.MustCompile(phoneRENamedCaps)
*/
func NamedFindStringSubmatch(s string, r *regexp.Regexp) (e2map.Map, bool) {
	rs := make(e2map.Map)
	if !r.MatchString(s) {
		return rs, false
	}
	match := r.FindStringSubmatch(s)
	for i, name := range r.SubexpNames() {
		rs[name] = match[i]
	}
	return rs, true
}
