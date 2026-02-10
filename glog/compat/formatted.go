package compat

import (
	"context"
	"fmt"
	"strings"

	"github.com/goliatone/go-logger/glog"
)

// FormattedLogger is a format-style logger contract used by legacy integrations.
type FormattedLogger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type formattedAdapter struct {
	logger FormattedLogger
}

type loggerAdapter struct {
	logger glog.Logger
}

// FromFormatted adapts a format-style logger into a glog.Logger.
func FromFormatted(logger FormattedLogger) glog.Logger {
	if logger == nil {
		return glog.Nop()
	}
	return &formattedAdapter{logger: logger}
}

// ToFormatted adapts a glog.Logger into a format-style logger.
func ToFormatted(logger glog.Logger) FormattedLogger {
	return &loggerAdapter{
		logger: glog.Ensure(logger),
	}
}

// PrintfFunc returns a printf-style callback bound to the requested level.
func PrintfFunc(logger glog.Logger, level string) func(format string, args ...any) {
	logger = glog.Ensure(logger)

	switch strings.ToLower(strings.TrimSpace(level)) {
	case strings.ToLower(glog.Trace):
		return func(format string, args ...any) {
			logger.Trace(fmt.Sprintf(format, args...))
		}
	case strings.ToLower(glog.Debug):
		return func(format string, args ...any) {
			logger.Debug(fmt.Sprintf(format, args...))
		}
	case strings.ToLower(glog.Warn), "warning":
		return func(format string, args ...any) {
			logger.Warn(fmt.Sprintf(format, args...))
		}
	case strings.ToLower(glog.Error), "err":
		return func(format string, args ...any) {
			logger.Error(fmt.Sprintf(format, args...))
		}
	case strings.ToLower(glog.Fatal):
		return func(format string, args ...any) {
			logger.Fatal(fmt.Sprintf(format, args...))
		}
	default:
		return func(format string, args ...any) {
			logger.Info(fmt.Sprintf(format, args...))
		}
	}
}

func (a *formattedAdapter) Trace(msg string, args ...any) {
	a.logger.Debugf(msg, args...)
}

func (a *formattedAdapter) Debug(msg string, args ...any) {
	a.logger.Debugf(msg, args...)
}

func (a *formattedAdapter) Info(msg string, args ...any) {
	a.logger.Infof(msg, args...)
}

func (a *formattedAdapter) Warn(msg string, args ...any) {
	a.logger.Warnf(msg, args...)
}

func (a *formattedAdapter) Error(msg string, args ...any) {
	a.logger.Errorf(msg, args...)
}

func (a *formattedAdapter) Fatal(msg string, args ...any) {
	a.logger.Errorf(msg, args...)
}

func (a *formattedAdapter) WithContext(ctx context.Context) glog.Logger {
	return a
}

func (a *loggerAdapter) Debugf(format string, args ...any) {
	a.logger.Debug(fmt.Sprintf(format, args...))
}

func (a *loggerAdapter) Infof(format string, args ...any) {
	a.logger.Info(fmt.Sprintf(format, args...))
}

func (a *loggerAdapter) Warnf(format string, args ...any) {
	a.logger.Warn(fmt.Sprintf(format, args...))
}

func (a *loggerAdapter) Errorf(format string, args ...any) {
	a.logger.Error(fmt.Sprintf(format, args...))
}
