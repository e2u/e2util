package e2var

import (
	"testing"
)

func TestMustStringValue(t *testing.T) {
	tests := []struct {
		input    *string
		expected string
	}{
		{nil, ""},
		{new("hello"), "hello"},
	}

	for _, tt := range tests {
		result := MustStringValue(tt.input)
		if result != tt.expected {
			t.Errorf("MustStringValue(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestNeverNullPoint(t *testing.T) {
	tests := []struct {
		input    any
		defVal   any
		expected any
	}{
		{nil, "default", "default"},
		{"", "default", "default"},
		{"hello", "default", "hello"},
		{0, 42, 42},
		{15, 42, 15},
		{0.0, 3.14, 3.14},
		{2.5, 3.14, 2.5},
		{false, true, true},
		{true, false, true},
	}

	for _, tt := range tests {
		switch v := tt.input.(type) {
		case string:
			result := *NeverNullPoint(v, tt.defVal.(string))
			if result != tt.expected {
				t.Errorf("NeverNullPoint(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		case int:
			result := *NeverNullPoint(v, tt.defVal.(int))
			if result != tt.expected {
				t.Errorf("NeverNullPoint(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		case float64:
			result := *NeverNullPoint(v, tt.defVal.(float64))
			if result != tt.expected {
				t.Errorf("NeverNullPoint(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		case bool:
			result := *NeverNullPoint(v, tt.defVal.(bool))
			if result != tt.expected {
				t.Errorf("NeverNullPoint(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		}
	}
}

func TestIfElse(t *testing.T) {
	if IfElse(1, 1, "yes", "no") != "yes" {
		t.Error("IfElse(1, 1) failed")
	}
	if IfElse(1, 2, "yes", "no") != "no" {
		t.Error("IfElse(1, 2) failed")
	}
}

func TestIfElseFunc(t *testing.T) {
	t.Run("v1 equals v2", func(t *testing.T) {
		var executed string
		IfElseFunc(1, 1,
			func() { executed = "f1" },
			func() { executed = "f2" })
		if executed != "f1" {
			t.Errorf("IfElseFunc(1, 1) executed %q, want %q", executed, "f1")
		}
	})

	t.Run("v1 not equals v2", func(t *testing.T) {
		var executed string
		IfElseFunc(1, 2,
			func() { executed = "f1" },
			func() { executed = "f2" })
		if executed != "f2" {
			t.Errorf("IfElseFunc(1, 2) executed %q, want %q", executed, "f2")
		}
	})
}

func TestTrueThen(t *testing.T) {
	if TrueThen(true, 1, 0) != 1 {
		t.Error("TrueThen(true) failed")
	}
	if TrueThen(false, 1, 0) != 0 {
		t.Error("TrueThen(false) failed")
	}
}

func TestNotNullThen(t *testing.T) {
	var s *string
	if NotNullThen(s, "yes", "no") != "no" {
		t.Error("NotNullThen(nil) failed")
	}
	str := "test"
	if NotNullThen(&str, "yes", "no") != "yes" {
		t.Error("NotNullThen(non-nil) failed")
	}
}

func TestNullThen(t *testing.T) {
	var s *string
	if NullThen(s, "yes", "no") != "yes" {
		t.Error("NullThen(nil) failed")
	}
	str := "test"
	if NullThen(&str, "yes", "no") != "no" {
		t.Error("NullThen(non-nil) failed")
	}
}

func TestValueOrDefault(t *testing.T) {
	tests := []struct {
		input    any
		defVal   any
		expected any
	}{
		{nil, "default", "default"},
		{"", "default", "default"},
		{"hello", "default", "hello"},
		{0, 42, 42},
		{15, 42, 15},
	}

	for _, tt := range tests {
		switch v := tt.input.(type) {
		case string:
			result := ValueOrDefault(v, tt.defVal.(string))
			if result != tt.expected {
				t.Errorf("ValueOrDefault(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		case int:
			result := ValueOrDefault(v, tt.defVal.(int))
			if result != tt.expected {
				t.Errorf("ValueOrDefault(%v, %v) = %v, want %v", v, tt.defVal, result, tt.expected)
			}
		}
	}
}

func TestExpectOrDefault(t *testing.T) {
	tests := []struct {
		input    any
		defVal   any
		expected any
		ok       bool
	}{
		{"hello", "default", "hello", true},
		{42, "default", "default", false},
		{nil, 0, 0, false},
	}

	for _, tt := range tests {
		switch def := tt.defVal.(type) {
		case string:
			result, ok := ExpectOrDefault(tt.input, def)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("ExpectOrDefault(%v, %v) = %v, %v, want %v, %v",
					tt.input, def, result, ok, tt.expected, tt.ok)
			}
		case int:
			result, ok := ExpectOrDefault(tt.input, def)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("ExpectOrDefault(%v, %v) = %v, %v, want %v, %v",
					tt.input, def, result, ok, tt.expected, tt.ok)
			}
		}
	}
}

func TestP(t *testing.T) {
	value := 42
	ptr := new(value)
	if *ptr != value {
		t.Errorf("P(%v) = %v, want %v", value, *ptr, value)
	}
}
