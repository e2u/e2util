package e2logrus

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// Config 日志配置
type Config struct {
	// Output 配置日志输出,有以下格式: file:///,stdout,stderr 等,文件路径设置例:
	// /opt/logs/log.%Y%m%d
	// 具体见  strftime(3) 格式
	Output                    string    `mapstructure:"output"`
	Level                     string    `mapstructure:"level"`         // 配置日志输出级别: trace,debug,info,warn,error
	MaxAge                    int       `mapstructure:"max_age"`       // 日志保留天数
	RotationTime              int       `mapstructure:"rotation_time"` // 日志分割时间,单位秒,默认86400秒
	Format                    string    `mapstructure:"format"`        // 格式, json | text, default: json
	DisableReportCaller       bool      `mapstructure:"disable_report_caller"`
	DisableColor              bool      `mapstructure:"disable_color"`
	EnvironmentOverrideColors bool      `mapstructure:"environment_override_colors"`
	DisableQuote              bool      `mapstructure:"disable_quote"`
	DisableFullTimestamp      bool      `mapstructure:"disable_full_timestamp"`
	DisableQuoteEmptyFields   bool      `mapstructure:"disable_quote_empty_fields"`
	DisablePadLevelText       bool      `mapstructure:"disable_pad_level_text"`
	PrettyPrint               bool      `mapstructure:"pretty_print"`        // json only
	DisableHTMLEscape         bool      `mapstructure:"disable_html_escape"` // json only
	DataKey                   string    `mapstructure:"data_key"`            // json only, when use WithField, the Fields  will inside the DataKey, default fields
	Writer                    io.Writer `mapstructure:"-"`
}

func CloneLogrus(orig *logrus.Logger) *logrus.Logger {
	newLogger := logrus.New()
	newLogger.SetFormatter(orig.Formatter)
	newLogger.SetOutput(orig.Out)
	newLogger.SetReportCaller(orig.ReportCaller)
	newLogger.SetLevel(orig.Level)
	return newLogger
}

//func DefaultConfig() *Config {
//	return &Config{
//		Output:       "stdout",
//		Level:        "debug",
//		MaxAge:       365,
//		RotationTime: 86400,
//		Format:       "json",
//	}
//}

// NewWriter 返回一个 writer,可以在 logrus 中使用
//func NewWriter(config *Config) (io.Writer, error) {
//	switch config.Output {
//	case "stdout":
//		return os.Stdout, nil
//	case "stderr":
//		return os.Stderr, nil
//	}
//
//	if !strings.HasPrefix(config.Output, `file://`) {
//		return os.Stdout, nil
//	}
//
//	logPath := strings.ReplaceAll(config.Output, `file://`, "")
//	linkName := filepath.Join(filepath.Dir(logPath), "current")
//
//	// 默认保留1年日志
//	if config.MaxAge == 0 {
//		config.MaxAge = 365
//	}
//
//	if config.RotationTime <= 0 {
//		config.RotationTime = 86400
//	}
//
//	rl, err := rotatelogs.New(logPath,
//		rotatelogs.WithMaxAge(24*time.Hour*time.Duration(config.MaxAge)),
//		rotatelogs.WithRotationTime(time.Second*time.Duration(config.RotationTime)),
//		rotatelogs.WithLinkName(linkName),
//	)
//
//	return rl, err
//}

var seqNum uint64

type SeqHook struct{}

func (h *SeqHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *SeqHook) Fire(entry *logrus.Entry) error {
	seq := atomic.AddUint64(&seqNum, 1)
	entry.Data["seq"] = fmt.Sprintf("0x%016x", seq)
	return nil
}
