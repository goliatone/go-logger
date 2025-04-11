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
	Success(msg string, args ...any)
	Error(msg string, args ...any)
	WithContext(ctx context.Context) Logger
}

type LoggerProvider interface {
	GetLogger(name string) Logger
}

const (
	LevelTrace = slog.Level(-8)
	// LevelDebug Level = -4
	// LevelInfo  Level = 0
	// LevelWarn  Level = 4
	LevelSuccess = slog.Level(6)
	// LevelError Level = 8
)

var CustomLevels = map[slog.Leveler]string{
	LevelTrace:   "TRACE",
	LevelSuccess: "SUCCESS",
}

const (
	LoggerTypeConsole = "console"
	LoggerTypePretty  = "pretty"
	LoggerTypeJSON    = "json"
)
