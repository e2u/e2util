package e2db

import (
	"strings"
	"time"

	"github.com/e2u/e2util/e2logrus"
	"github.com/sirupsen/logrus"
	gormlogger "gorm.io/gorm/logger"
)

func newLogger(cfg *Config) gormlogger.Interface {
	if cfg.SQLLogSlowThreshold <= 0 {
		cfg.SQLLogSlowThreshold = 200
	}
	ll := gormlogger.Info
	var dbLogger *logrus.Logger // dbLogger := e2logrus.CloneLogrus(logrus.StandardLogger())
	if cfg.LoggerConfig == nil {
		dbLogger = e2logrus.CloneLogrus(logrus.StandardLogger())
	} else {
		dbLogger = e2logrus.NewLogger(cfg.LoggerConfig)
		ll = func(lvs string) gormlogger.LogLevel {
			switch strings.ToLower(lvs) {
			case "silent":
				return gormlogger.Silent
			case "debug", "info":
				return gormlogger.Info
			case "warn", "warning":
				return gormlogger.Warn
			case "error", "err":
				return gormlogger.Error
			default:
				return gormlogger.Info
			}
		}(cfg.LoggerConfig.Level)
	}

	dbLogger.AddHook(&e2logrus.SeqHook{})
	SQLLogColorful := cfg.SQLLogColorful
	if cfg.LoggerConfig.Format == "json" {
		SQLLogColorful = false
	}

	return gormlogger.New(dbLogger, gormlogger.Config{
		SlowThreshold:             time.Duration(cfg.SQLLogSlowThreshold) * time.Millisecond,
		LogLevel:                  ll,
		IgnoreRecordNotFoundError: cfg.SQLLogIgnoreRecordNotFoundError,
		Colorful:                  SQLLogColorful,
	})
}
