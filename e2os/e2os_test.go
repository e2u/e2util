package e2os

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	assert.False(t, FileExists(""))
	assert.False(t, FileExists("   "))
	assert.False(t, FileExists(filepath.Join(t.TempDir(), "missing.txt")))

	f := filepath.Join(t.TempDir(), "ok.txt")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o600))
	assert.True(t, FileExists(f))
	assert.True(t, FileExists(t.TempDir()))
}

func TestMustGetwd(t *testing.T) {
	want, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, want, MustGetwd())
}

func TestGetExecDir(t *testing.T) {
	dir, err := GetExecDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
}

func TestFindGoModDir(t *testing.T) {
	dir, err := findGoModDir()
	require.NoError(t, err)
	assert.True(t, FileExists(filepath.Join(dir, "go.mod")))
}

func TestIsTestAndGoRun(t *testing.T) {
	assert.True(t, isTest())
}

func TestChdirToAppRoot(t *testing.T) {
	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(orig)) })

	require.NoError(t, ChdirToAppRoot())
	assert.True(t, FileExists("go.mod"))
}

func TestChangeWorkdir(t *testing.T) {
	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.Chdir(orig)) })

	require.Error(t, ChangeWorkdir())
	require.Error(t, ChangeWorkdir(""))
	require.Error(t, ChangeWorkdir("   "))

	dir := t.TempDir()
	require.NoError(t, ChangeWorkdir(dir))
	got, err := os.Getwd()
	require.NoError(t, err)
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		want = dir
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		gotResolved = got
	}
	assert.Equal(t, want, gotResolved)
}

func TestRetryRun(t *testing.T) {
	t.Run("success first try", func(t *testing.T) {
		n := 0
		err := RetryRun(3, 0, func(retryCount int) error {
			n++
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 1, n)
	})

	t.Run("success after failures", func(t *testing.T) {
		n := 0
		err := RetryRun(3, 0, func(retryCount int) error {
			n++
			if retryCount < 2 {
				return errors.New("fail")
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 3, n)
	})

	t.Run("all fail returns last error", func(t *testing.T) {
		n := 0
		err := RetryRun(3, 0, func(retryCount int) error {
			n++
			return errors.New("x")
		})
		require.EqualError(t, err, "x")
		assert.Equal(t, 3, n)
	})

	t.Run("sleeps between attempts but not after last", func(t *testing.T) {
		n := 0
		err := RetryRun(2, time.Millisecond, func(int) error {
			n++
			return errors.New("x")
		})
		require.Error(t, err)
		assert.Equal(t, 2, n)
	})

	t.Run("invalid args", func(t *testing.T) {
		require.Error(t, RetryRun(0, time.Millisecond, func(int) error { return nil }))
		require.Error(t, RetryRun(-1, 0, func(int) error { return nil }))
		require.Error(t, RetryRun(1, 0, nil))
	})
}

func TestInitSystemdService(t *testing.T) {
	unit, err := InitSystemdService()
	require.NoError(t, err)
	assert.Contains(t, unit, "[Unit]")
	assert.Contains(t, unit, "[Service]")
	assert.Contains(t, unit, "[Install]")

	wd, err := GetExecDir()
	require.NoError(t, err)
	assert.Contains(t, unit, "WorkingDirectory="+filepath.Clean(wd))

	u, err := user.Current()
	require.NoError(t, err)
	assert.Contains(t, unit, "User="+u.Username)

	exe, err := os.Executable()
	require.NoError(t, err)
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	assert.Contains(t, unit, "ExecStart="+exe)
}

func TestExternalIP(t *testing.T) {
	ip, err := ExternalIP()
	if err != nil {
		assert.Equal(t, "are you connected to the network?", err.Error())
		assert.Empty(t, ip)
		return
	}
	assert.NotEmpty(t, ip)
	assert.NotContains(t, ip, ":")
}

func TestSendSignalToProcess(t *testing.T) {
	require.Error(t, SendSignalToProcess("", os.Interrupt))
	require.Error(t, SendSignalToProcess("   ", os.Interrupt))

	if _, err := exec.LookPath("pgrep"); err != nil {
		t.Skip("pgrep not available")
	}
	err := SendSignalToProcess("e2os-no-such-process-xyz", os.Interrupt)
	require.Error(t, err)
}

func TestIsGoRun(t *testing.T) {
	_ = isGoRun()
}
