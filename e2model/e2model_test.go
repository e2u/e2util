package e2model

import (
	"errors"
	"testing"
)

func TestHttpPatchAllow(t *testing.T) {
	h := &HttpPatch{Op: HttpPatchOpReplace, Path: "name"}
	if !h.AllowOp([]string{HttpPatchOpAdd, HttpPatchOpReplace}) {
		t.Fatal("expected replace to be allowed")
	}
	if h.AllowOp([]string{HttpPatchOpRemove}) {
		t.Fatal("expected replace to be rejected")
	}
	if !h.AllowPath([]string{"name", "email"}) {
		t.Fatal("expected path name to be allowed")
	}
	if h.AllowPath([]string{"email"}) {
		t.Fatal("expected path name to be rejected")
	}
}

func TestNewNullBool(t *testing.T) {
	ok := NewNullBool(true, nil)
	if !ok.Bool || !ok.Valid || ok.Error != nil {
		t.Errorf("NewNullBool(true, nil) = %+v", ok)
	}
	err := errors.New("boom")
	bad := NewNullBool(false, err)
	if bad.Valid || bad.Error != err {
		t.Errorf("NewNullBool(false, err) = %+v", bad)
	}
}
