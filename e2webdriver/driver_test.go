package e2webdriver

import (
	"context"
	"os"
	"testing"

	"github.com/e2u/e2util/e2exec"
)

var ctx = context.TODO()

func skipUnlessWebdriver(t *testing.T) {
	t.Helper()
	if os.Getenv("E2UTIL_WEBDRIVER_TEST") == "" {
		t.Skip("set E2UTIL_WEBDRIVER_TEST=1 to run chromedriver download tests")
	}
}

func Test_getVersions(t *testing.T) {
	skipUnlessWebdriver(t)
	t.Log(getLatestVersions(ctx))
}

func Test_buildDownloadUrl(t *testing.T) {
	skipUnlessWebdriver(t)
	ver := e2exec.Must(getLatestVersions(ctx))
	u, err := buildDownloadUrl(ctx, ver)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := downloadAndUnzip(ctx, u, t.TempDir())
	t.Log(paths)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Install(t *testing.T) {
	skipUnlessWebdriver(t)
	exePath, err := Install(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(exePath)
}
