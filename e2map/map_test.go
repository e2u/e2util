package e2map

import (
	"testing"
)

func Test_Map(t *testing.T) {
	m := make(Map)
	m["a"] = 1
	m["b"] = 2
	m["c"] = "Hello"
	m["d"] = true

	if v, ok := m["a"]; !ok || v != 1 {
		t.Fatal()
	}

	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Fatal()
	}

	if v, ok := m.DefaultGet("a", 2); !ok || v != 1 {
		t.Fatal()
	}
}
