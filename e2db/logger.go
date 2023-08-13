package e2db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	adapterLogrus = "logurs"
	adapterSlog   = "slog"
)

type logger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	gormlogger.Config
	adapter string
}

func NewLogger(level string, adapter string) *logger {

	ll := func(lvs string) gormlogger.LogLevel {
		switch strings.ToLower(lvs) {
		case "Silent":
			return gormlogger.Silent
		case "Info":
			return gormlogger.Info
		case "Warn":
			return gormlogger.Warn
		case "Error":
			return gormlogger.Error
		default:
			return gormlogger.Info
		}
	}(level)

	if adapter == "" {
		adapter = adapterLogrus
	}

	l := &logger{
		adapter:               adapter,
		SkipErrRecordNotFound: true,
	}

	l.LogLevel = ll
	return l
}

func (l *logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *logger) Info(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		if l.adapter == adapterLogrus {
			log.WithContext(ctx).Infof(s, args...)
		} else {
			slog.InfoContext(ctx, s, args...)
		}
	}
}

func (l *logger) Warn(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		if l.adapter == adapterLogrus {
			log.WithContext(ctx).Warnf(s, args...)
		} else {
			slog.WarnContext(ctx, s, args...)
		}
	}
}

func (l *logger) Error(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		if l.adapter == adapterLogrus {
			log.WithContext(ctx).Errorf(s, args...)
		} else {
			slog.ErrorContext(ctx, s, args...)
		}
	}
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[log.ErrorKey] = err
		if l.adapter == adapterLogrus {
			log.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		} else {
			slog.ErrorContext(ctx, fmt.Sprintf("%s [%s]", sql, elapsed), "fields", fields)
		}
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		if l.adapter == adapterLogrus {
			log.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		} else {
			slog.WarnContext(ctx, fmt.Sprintf("%s [%s]", sql, elapsed), "fields", fields)
		}
		return
	}

	if l.adapter == adapterLogrus {
		log.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
	} else {
		slog.DebugContext(ctx, fmt.Sprintf("%s [%s]", sql, elapsed), "fields", fields)
	}

}
