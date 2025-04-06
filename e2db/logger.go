package e2db

import (
	"strings"
	"time"

	"github.com/e2u/e2util/e2logrus"
	"github.com/sirupsen/logrus"
	gormlogger "gorm.io/gorm/logger"
)

const defaultSlowThreshold = 200 // ms

var logLevelMap = map[string]gormlogger.LogLevel{
	"silent":  gormlogger.Silent,
	"debug":   gormlogger.Info,
	"info":    gormlogger.Info,
	"warn":    gormlogger.Warn,
	"warning": gormlogger.Warn,
	"error":   gormlogger.Error,
	"err":     gormlogger.Error,
}

func newLogger(cfg *Config) gormlogger.Interface {
	if cfg.SQLLogSlowThreshold <= 0 {
		cfg.SQLLogSlowThreshold = defaultSlowThreshold
	}

	var dbLogger *logrus.Logger
	logLevel := gormlogger.Info
	if cfg.LoggerConfig == nil {
		dbLogger = e2logrus.CloneLogrus(logrus.StandardLogger())
	} else {
		dbLogger = e2logrus.NewLogger(cfg.LoggerConfig)
		if level, ok := logLevelMap[strings.ToLower(cfg.LoggerConfig.Level)]; ok {
			logLevel = level
		} else {
			dbLogger.Warnf("unknown log level '%s', defaulting to 'info'", cfg.LoggerConfig.Level)
		}
	}

	dbLogger.AddHook(&e2logrus.SeqHook{})
	isJSONFormat := cfg.LoggerConfig != nil && cfg.LoggerConfig.Format == "json"
	colorful := cfg.SQLLogColorful && !isJSONFormat

	return gormlogger.New(dbLogger, gormlogger.Config{
		SlowThreshold:             time.Duration(cfg.SQLLogSlowThreshold) * time.Millisecond,
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: cfg.SQLLogIgnoreRecordNotFoundError,
		Colorful:                  colorful,
	})
}
