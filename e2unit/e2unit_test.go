package e2unit

import "testing"

func TestBytes(t *testing.T) {
	if got := Bytes(82854982); got != "83 MB" {
		t.Errorf("Bytes(82854982) = %q, want 83 MB", got)
	}
	if got := Bytes(5); got != "5 B" {
		t.Errorf("Bytes(5) = %q, want 5 B", got)
	}
}

func TestIBytes(t *testing.T) {
	if got := IBytes(82854982); got != "79 MiB" {
		t.Errorf("IBytes(82854982) = %q, want 79 MiB", got)
	}
}

func TestParseBytes(t *testing.T) {
	got, err := ParseBytes("42MB")
	if err != nil {
		t.Fatal(err)
	}
	if got != 42000000 {
		t.Errorf("ParseBytes(42MB) = %d, want 42000000", got)
	}

	got, err = ParseBytes("42mib")
	if err != nil {
		t.Fatal(err)
	}
	if got != 44040192 {
		t.Errorf("ParseBytes(42mib) = %d, want 44040192", got)
	}

	if _, err := ParseBytes("12xx"); err == nil {
		t.Fatal("expected error for unknown unit")
	}
}
