package e2os

import (
	"os"
	"path/filepath"
	"strings"
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
		return "", nil
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
