package glog

import (
	"context"
	"log/slog"
)

type Logger interface {
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	WithContext(ctx context.Context) Logger
}

type LoggerProvider interface {
	GetLogger(name string) Logger
}

const (
	LevelTrace = slog.Level(-8)
)

const (
	Error   = "ERROR"
	Success = "SUCCESS"
	Warn    = "WARN"
	Info    = "INFO"
	Debug   = "DEBUG"
	Trace   = "TRACE"
)

var CustomLevels = map[slog.Leveler]string{
	LevelTrace: Trace,
}

const (
	LoggerTypeConsole = "console"
	LoggerTypePretty  = "pretty"
	LoggerTypeJSON    = "json"
)
