package e2os

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/e2u/e2util/e2strconv"
	"github.com/sirupsen/logrus"
)

// ChdirToAppRoot 根據執行模式切換工作目錄：
// - go run / go test → 切到 go.mod 所在目錄（module root）
// - 編譯後的 binary → 切到 binary 所在目錄
func ChdirToAppRoot() error {
	if isTest() || isGoRun() {
		dir, err := findGoModDir()
		if err != nil {
			return fmt.Errorf("find go.mod: %w", err)
		}
		return os.Chdir(dir)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("eval symlinks: %w", err)
	}
	return os.Chdir(filepath.Dir(exe))
}

func isTest() bool {
	return testing.Testing()
}

func isGoRun() bool {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	return strings.Contains(exe, "go-build")
}

func findGoModDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}

func MustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

func GetExecDir() (string, error) {
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

func FileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && st != nil
}

func SendSignalToProcess(processName string, signal os.Signal) error {
	processName = strings.TrimSpace(processName)
	if processName == "" {
		return fmt.Errorf("process name is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pgrep", processName) // #nosec G204
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("execute command error=%v command=pgrep %s", err, processName)
		return fmt.Errorf("pgrep %s: %w", processName, err)
	}

	signaled := 0
	for _, pidStr := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}
		pid, err := e2strconv.ParseInt[int](pidStr)
		if err != nil {
			return fmt.Errorf("parse pid %q: %w", pidStr, err)
		}
		if pid <= 0 {
			return fmt.Errorf("invalid pid %d", pid)
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			logrus.Errorf("find process error=%v pid=%d", err, pid)
			return err
		}
		if err := process.Signal(signal); err != nil {
			logrus.Errorf("sending signal error=%v pid=%d", err, pid)
			return err
		}
		signaled++
	}
	if signaled == 0 {
		return fmt.Errorf("no process found for %s", processName)
	}
	return nil
}

func RetryRun(maxRetry int, sleep time.Duration, fn func(retryCount int) error) error {
	if maxRetry <= 0 {
		return fmt.Errorf("maxRetry must be positive")
	}
	if fn == nil {
		return fmt.Errorf("fn must not be nil")
	}
	var lastErr error
	for i := range maxRetry {
		if err := fn(i); err != nil {
			lastErr = err
			if i < maxRetry-1 && sleep > 0 {
				time.Sleep(sleep)
			}
			continue
		}
		return nil
	}
	return lastErr
}

func ChangeWorkdir(dir ...string) error {
	if len(dir) == 0 || strings.TrimSpace(dir[0]) == "" {
		return fmt.Errorf("workdir is required")
	}
	return os.Chdir(dir[0])
}

func InitSystemdService() (string, error) {
	const tmplStr = `[Unit]
Description={{.ServiceName}} daemon
After=network.target

[Service]
# Type: dbus, exec, forking, idle, notify, oneshot, simple
Type=simple

##
# Environment=VAR1=value1; VAR2=value2; VAR3=value3
# Or
#Environment=VAR1=value1
#Environment=VAR2=value2
#Environment=VAR3=value3
##
Environment="TZ={{.TimeZone}}"

WorkingDirectory={{.WorkingDirectory}}
ExecStart={{.Executable}} --env=prod

# Restart: always, no, on-abnormal, on-abort, on-failure, on-success, on-watchdog,
Restart=on-success
RestartSec=60

# KillMode: control-group, mixed, none, process
KillMode=process

# ExitType: main, cgroup
ExitType=main

TimeoutSec=300
TimeoutStartSec=300
TimeoutStopSec=300

User={{.USER}}
Group={{.GROUP}}


[Install]
WantedBy=multi-user.target
`

	exe, err := os.Executable()
	if err != nil {
		logrus.Errorf("get executable error=%v", err)
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	wd, err := GetExecDir()
	if err != nil {
		logrus.Errorf("get executable dir error=%v", err)
		return "", err
	}
	wd = filepath.Clean(wd)

	currentTimeZone, _ := time.Now().Zone()

	u, err := user.Current()
	if err != nil {
		logrus.Errorf("get current user error=%v", err)
		return "", err
	}
	group := ""
	if g, err := user.LookupGroupId(u.Gid); err == nil {
		group = g.Name
	}

	data := map[string]string{
		"ServiceName":      filepath.Base(exe),
		"TimeZone":         currentTimeZone,
		"WorkingDirectory": wd,
		"Executable":       exe,
		"USER":             u.Username,
		"GROUP":            group,
	}

	var buf bytes.Buffer
	tmpl, err := template.New("systemd").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		logrus.Errorf("execute template error=%v", err)
		return "", err
	}
	return buf.String(), nil
}
