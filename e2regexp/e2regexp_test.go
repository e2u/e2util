package e2regexp

import (
	"regexp"
	"testing"
)

func TestNamedFindStringSubmatch(t *testing.T) {
	re := regexp.MustCompile(`(?P<area>\d{3})\-(?P<exchange>\d{3})\-(?P<line>\d{4})$`)
	got, ok := NamedFindStringSubmatch("202-555-0147", re)
	if !ok {
		t.Fatal("expected match")
	}
	if got["area"] != "202" || got["exchange"] != "555" || got["line"] != "0147" {
		t.Errorf("got %#v", got)
	}

	if _, ok := NamedFindStringSubmatch("not-a-phone", re); ok {
		t.Fatal("expected no match")
	}
}
