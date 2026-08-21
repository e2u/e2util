package e2io

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMustReadAll(t *testing.T) {
	got := MustReadAll(strings.NewReader("hello"))
	if string(got) != "hello" {
		t.Errorf("MustReadAll = %q", got)
	}
	if s := MustReadAllAsString(strings.NewReader("world")); s != "world" {
		t.Errorf("MustReadAllAsString = %q", s)
	}
}

func TestMustReadAllAndClose(t *testing.T) {
	rc := io.NopCloser(strings.NewReader("abc"))
	got := MustReadAllAndClose(rc)
	if string(got) != "abc" {
		t.Errorf("MustReadAllAndClose = %q", got)
	}
	if s := MustReadAllAsStringAndClose(io.NopCloser(strings.NewReader("xyz"))); s != "xyz" {
		t.Errorf("MustReadAllAsStringAndClose = %q", s)
	}
}

func TestMustReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("file-body"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := MustReadFile(path); string(got) != "file-body" {
		t.Errorf("MustReadFile = %q", got)
	}
	if got := MustReadFile(filepath.Join(dir, "missing.txt")); got != nil {
		t.Errorf("MustReadFile(missing) = %q, want nil", got)
	}
}
