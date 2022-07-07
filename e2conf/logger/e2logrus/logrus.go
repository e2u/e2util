package e2logrus

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// Config 日志配置
type Config struct {
	// Output 配置日志输出,有以下格式: file:///,stdout,stderr 等,文件路径设置例:
	// /opt/logs/log.%Y%m%d
	// 具体见  strftime(3) 格式
	Output       string
	LogLevel     string // 配置日志输出级别: trace,debug,info,warn,error
	MaxAge       int    // 日志保留天数
	RotationTime int    // 日志分割时间,单位秒,默认86400秒
}

// NewWriter 返回一个 writer,可以在 logrus 中使用
func NewWriter(config *Config) (io.Writer, error) {
	switch config.Output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	}

	// TODO 除了标准输出外，先支持文件输出
	if !strings.HasPrefix(config.Output, `file://`) {
		return os.Stdout, nil
	}

	logPath := strings.ReplaceAll(config.Output, `file://`, "")
	linkName := filepath.Join(filepath.Dir(logPath), "current")

	// 默认保留1年日志
	if config.MaxAge == 0 {
		config.MaxAge = 365
	}

	if config.RotationTime <= 0 {
		config.RotationTime = 86400
	}

	rl, err := rotatelogs.New(logPath,
		rotatelogs.WithMaxAge(24*time.Hour*time.Duration(config.MaxAge)),
		rotatelogs.WithRotationTime(time.Second*time.Duration(config.RotationTime)),
		rotatelogs.WithLinkName(linkName),
	)

	return rl, err
}
