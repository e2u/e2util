package e2os

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/e2u/e2util/e2strconv"
)

// GetRunDir 获取当前运行的目录
func GetRunDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	if !strings.Contains(filepath.Dir(ex), "go-build") {
		return filepath.Dir(ex), nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func SendSignalToProcess(processName string, signal os.Signal) error {
	getPidCmd := []string{"pgrep", processName}
	cmd := exec.Command(getPidCmd[0], getPidCmd[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("execute command", "error", err, "command", getPidCmd)
		return err
	}
	for _, pid := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		process, err := os.FindProcess(e2strconv.MustParseInt(pid))
		if err != nil {
			slog.Error("find process", "error", err, "pid", pid)
			return err
		}
		if err := process.Signal(signal); err != nil {
			slog.Error("sending signal", "error", err, "pid", pid)
			return err
		}
	}
	return nil
}

func RetryRun(maxRetry int, sleep time.Duration, fn func(retryCount int) error) error {
	var lastErr error
	for i := 0; i < maxRetry; i++ {
		if err := fn(i); err != nil {
			lastErr = err
			time.Sleep(sleep)
		} else {
			return nil
		}
	}
	return lastErr
}
