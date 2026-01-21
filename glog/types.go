package glog

import (
	"context"
	"log/slog"
)

type coder interface {
	Code() int
}

type statuser interface {
	Status() string
}

type Logger interface {
	Trace(msg string, args ...any)
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
	WithContext(ctx context.Context) Logger
}

type LoggerProvider interface {
	GetLogger(name string) Logger
}

// FieldsLogger is an optional extension for attaching structured fields.
type FieldsLogger interface {
	WithFields(map[string]any) Logger
}

type FatalBehavior int

const (
	FatalBehaviorExit FatalBehavior = iota
	FatalBehaviorPanic
	FatalBehaviorLogOnly
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(20)
)

const (
	Trace = "TRACE"
	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

var CustomLevels = map[slog.Leveler]string{
	LevelTrace: Trace,
	LevelFatal: Fatal,
}

const (
	LoggerTypeConsole = "console"
	LoggerTypePretty  = "pretty"
	LoggerTypeJSON    = "json"
)
