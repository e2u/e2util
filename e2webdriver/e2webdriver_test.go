package e2webdriver

import (
	"runtime"
	"testing"
)

func TestGetCurrentOSFileName(t *testing.T) {
	name, err := getCurrentOSFileName()
	key := runtime.GOOS + "_" + runtime.GOARCH
	if _, ok := fileNameMap[key]; !ok {
		if err == nil {
			t.Fatalf("expected error for unsupported platform %s", key)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || name != fileNameMap[key] {
		t.Errorf("got %q, want %q", name, fileNameMap[key])
	}
}
