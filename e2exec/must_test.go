package e2exec

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func Test_Must(t *testing.T) {
	f1 := func() (string, error) {
		return "abc", fmt.Errorf("error message")
	}
	f2 := func() (int, error) {
		return 100, fmt.Errorf("error message")
	}
	f3 := func() (float64, error) {
		return 100.05, fmt.Errorf("error message")
	}
	f4 := func() (bool, error) {
		return true, fmt.Errorf("error message")
	}

	if got := Must(f1()); got != "abc" {
		t.Fatalf("Must(string) = %q, want abc", got)
	}
	if got := Must(f2()); got != 100 {
		t.Fatalf("Must(int) = %d, want 100", got)
	}
	if got := Must(f3()); got != 100.05 {
		t.Fatalf("Must(float64) = %v, want 100.05", got)
	}
	if got := Must(f4()); !got {
		t.Fatalf("Must(bool) = %v, want true", got)
	}
}

func Test_MustNoError(t *testing.T) {
	if got := Must("ok", error(nil)); got != "ok" {
		t.Fatalf("Must() = %q, want ok", got)
	}
}

func Test_Must2(t *testing.T) {
	f1 := func() (string, string, error) {
		return "abc", "ABC", fmt.Errorf("error message")
	}
	f2 := func() (int, int, error) {
		return 100, 200, fmt.Errorf("error message")
	}
	f3 := func() (float64, string, error) {
		return 100.05, "hello", fmt.Errorf("error message")
	}
	f4 := func() (bool, int, error) {
		return true, 200, fmt.Errorf("error message")
	}

	if v1, v2 := Must2(f1()); v1 != "abc" || v2 != "ABC" {
		t.Fatalf("Must2(string) = %q, %q", v1, v2)
	}
	if v1, v2 := Must2(f2()); v1 != 100 || v2 != 200 {
		t.Fatalf("Must2(int) = %d, %d", v1, v2)
	}
	if v1, v2 := Must2(f3()); v1 != 100.05 || v2 != "hello" {
		t.Fatalf("Must2(float64) = %v, %q", v1, v2)
	}
	if v1, v2 := Must2(f4()); !v1 || v2 != 200 {
		t.Fatalf("Must2(bool) = %v, %d", v1, v2)
	}
}

type testCloser struct{ err error }

func (c testCloser) Close() error { return c.err }

func TestMustClose(t *testing.T) {
	MustClose(testCloser{})
	MustClose(testCloser{err: io.EOF})
}

func TestSilentError(t *testing.T) {
	SilentError()
	SilentError(nil)
	SilentError(errors.New("boom"))
	SilentError("value", errors.New("boom"))
	SilentError("value", nil)
	SilentError("not-an-error")
}

func TestOnlyError(t *testing.T) {
	if err := OnlyError(); err != nil {
		t.Fatalf("OnlyError() = %v, want nil", err)
	}
	want := errors.New("boom")
	if err := OnlyError(want); err != want {
		t.Fatalf("OnlyError(err) = %v, want boom", err)
	}
	if err := OnlyError("ok", want); err != want {
		t.Fatalf("OnlyError(ok, err) = %v, want boom", err)
	}
	if err := OnlyError("ok", nil); err != nil {
		t.Fatalf("OnlyError(ok, nil) = %v, want nil", err)
	}
}

func TestTrueThen(t *testing.T) {
	var got string
	TrueThen(true, func() { got = "t" }, func() { got = "f" })()
	if got != "t" {
		t.Fatalf("TrueThen(true) ran %q", got)
	}
	TrueThen(false, func() { got = "t" }, func() { got = "f" })()
	if got != "f" {
		t.Fatalf("TrueThen(false) ran %q", got)
	}

	if v := TrueThenExec(true, func() any { return 1 }, func() any { return 2 }); v != 1 {
		t.Fatalf("TrueThenExec(true) = %v", v)
	}
	if v := TrueThenExec(false, func() any { return 1 }, func() any { return 2 }); v != 2 {
		t.Fatalf("TrueThenExec(false) = %v", v)
	}

	got = ""
	TrueThenFunc(true, func() { got = "t" }, func() { got = "f" })
	if got != "t" {
		t.Fatalf("TrueThenFunc(true) ran %q", got)
	}

	var ptr *string
	got = ""
	NotNullThenFunc(ptr, func() { got = "nn" }, func() { got = "n" })
	if got != "n" {
		t.Fatalf("NotNullThenFunc(nil) ran %q", got)
	}
	s := "x"
	NullThenFunc(&s, func() { got = "n" }, func() { got = "nn" })
	if got != "nn" {
		t.Fatalf("NullThenFunc(non-nil) ran %q", got)
	}
}
