package e2os

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/e2u/e2util/e2strconv"
	"github.com/sirupsen/logrus"
)

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

func ChangeWorkdir(dir ...string) error {
	if len(dir) > 0 {
		return os.Chdir(dir[0])
	}

	return nil
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
	exeDir, err := GetExecDir()
	if err != nil {
		logrus.Errorf("get executable error=%v", err)
		return "", err
	}
	exe = filepath.Clean(filepath.Base(exe))
	wd := filepath.Clean(filepath.Dir(exeDir))

	currentTime := time.Now()
	currentTimeZone, _ := currentTime.Zone()

	u, err := user.Current()
	if err != nil {
		logrus.Errorf("get current user error=%v", err)
		return "", err
	}

	data := map[string]string{
		"ServiceName":      exe,
		"TimeZone":         currentTimeZone,
		"WorkingDirectory": wd,
		"Executable":       filepath.Join(wd, exe),
		"USER":             u.Name,
		"GROUP":            "",
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("systemd").Parse(tmplStr))
	tmpl.Execute(&buf, data)
	return "", nil
}
