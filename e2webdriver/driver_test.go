package e2webdriver

import (
	"context"
	"testing"

	"github.com/e2u/e2util/e2exec"
)

var ctx = context.TODO()

func Test_getVersions(t *testing.T) {
	t.Log(getLatestVersions(ctx))
}

func Test_buildDownloadUrl(t *testing.T) {
	ver := e2exec.Must(getLatestVersions(ctx))
	u, err := buildDownloadUrl(ctx, ver)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := downloadAndUnzip(ctx, u, "/Volumes/r1/")
	t.Log(paths)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Install(t *testing.T) {
	exePath, err := Install(ctx, "/Volumes/r1")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(exePath)
}
