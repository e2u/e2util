package e2logrus

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/e2u/e2util/e2var"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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

func defaultConfig() *Config {
	return &Config{
		Output:                    "stdout",
		Level:                     "level",
		MaxAge:                    365,
		RotationTime:              86400,
		Format:                    "text",
		DisableReportCaller:       false,
		DisableColor:              false,
		EnvironmentOverrideColors: true,
		DisableQuote:              false,
		DisableFullTimestamp:      false,
		DisableQuoteEmptyFields:   false,
		DisablePadLevelText:       false,
		PrettyPrint:               false,
		DisableHTMLEscape:         false,
	}
}

func CloneLogrus(orig *logrus.Logger) *logrus.Logger {
	newLogger := logrus.New()
	newLogger.SetFormatter(orig.Formatter)
	newLogger.SetOutput(orig.Out)
	newLogger.SetReportCaller(orig.ReportCaller)
	newLogger.SetLevel(orig.Level)
	return newLogger
}

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

func NewLogger(cfg *Config) *logrus.Logger {
	cfg = e2var.NullThen(cfg, defaultConfig(), cfg)

	log := logrus.New()
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339Nano,
			DisableTimestamp:  cfg.DisableFullTimestamp,
			DisableHTMLEscape: cfg.DisableHTMLEscape,
			DataKey:           e2var.ValueOrDefault(cfg.DataKey, "fields"),
			PrettyPrint:       cfg.PrettyPrint,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:               !cfg.DisableColor, //
			DisableColors:             cfg.DisableColor,  //
			ForceQuote:                !cfg.DisableQuote,
			DisableQuote:              cfg.DisableQuote,
			EnvironmentOverrideColors: cfg.EnvironmentOverrideColors, //
			DisableTimestamp:          cfg.DisableFullTimestamp,
			FullTimestamp:             !cfg.DisableFullTimestamp,
			TimestampFormat:           time.RFC3339Nano,
			PadLevelText:              !cfg.DisablePadLevelText,
			QuoteEmptyFields:          !cfg.DisableQuoteEmptyFields, //
		})
	}

	log.SetReportCaller(!cfg.DisableReportCaller)

	rotationTime := cfg.RotationTime
	if rotationTime <= 0 {
		rotationTime = 86400
	}
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 365
	}

	logLevelStr := cfg.Level
	if logLevelStr == "" {
		logLevelStr = logrus.InfoLevel.String()
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	output := cfg.Output
	if output == "" {
		output = "stdout"
	}
	if strings.HasPrefix(output, "file://") {
		logPath := output[7:]
		linkName := filepath.Join(filepath.Dir(logPath), "current")
		fileName := filepath.Base(logPath)
		if fileName != "" {
			re := regexp.MustCompile(`%[a-zA-Z]+`)
			fileName = re.ReplaceAllString(fileName, "")
			fileName = strings.ReplaceAll(fileName, ".", "")
			fileName = strings.ReplaceAll(fileName, "-", "")
			fileName = strings.ReplaceAll(fileName, " ", "")
			linkName = filepath.Join(filepath.Dir(logPath), fileName+"-current")
		}

		rl, err := rotatelogs.New(logPath,
			rotatelogs.WithMaxAge(24*time.Hour*time.Duration(maxAge)),
			rotatelogs.WithRotationTime(time.Second*time.Duration(rotationTime)),
			rotatelogs.WithLinkName(linkName),
		)

		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(rl)
	} else {
		log.SetOutput(os.Stdout)
	}
	log.AddHook(&SeqHook{})

	log.Trace("active logrus TRACE level")
	log.Debug("active logrus DEBUG level")
	log.Info("active logrus INFO level")
	log.Warn("active logrus WARN level")
	log.Error("active logrus ERROR level")
	return log
}
