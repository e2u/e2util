package e2exec

import (
	"fmt"
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

	if Must(f1()) != "abc" {
		t.Fail()
	}

	if Must(f2()) != 100 {
		t.Fail()
	}

	if Must(f3()) >= 100.05 {
		t.Fail()
	}
	if Must(f4()) {
		t.Fail()
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

	if v1, v2 := Must2(f1()); v1 != "abc" && v2 != "AB" {
		t.Fail()
	}

	if v1, v2 := Must2(f2()); v1 != 100 && v2 != 200 {
		t.Fail()
	}

	if v1, v2 := Must2(f3()); v1 >= 100.05 && v2 != "hello" {
		t.Fail()
	}

	if v1, v2 := Must2(f4()); v1 && v2 == 200 {
		t.Fail()
	}
}
