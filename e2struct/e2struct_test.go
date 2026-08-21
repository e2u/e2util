package e2struct

import "testing"

type nested struct {
	Name string
}

type sample struct {
	Title  string
	Nested *nested
	Inner  nested
}

func TestPrepareStruct(t *testing.T) {
	s := &sample{Title: "  hello  "}
	PrepareStruct(s)
	if s.Title != "hello" {
		t.Errorf("Title = %q, want hello", s.Title)
	}
	if s.Nested == nil {
		t.Fatal("expected Nested to be initialized")
	}
	s.Nested.Name = "  inner  "
	PrepareStruct(s)
	if s.Nested.Name != "inner" {
		t.Errorf("Nested.Name = %q, want inner", s.Nested.Name)
	}
}

func TestPrepareStructIgnoresNonPointer(t *testing.T) {
	PrepareStruct(sample{Title: "  x  "})
	PrepareStruct((*sample)(nil))
	PrepareStruct("not a struct")
}
