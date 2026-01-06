package glog

import (
	"context"
	"io"
	"log/slog"
)

type Option func(*BaseLogger)

func WithLevel(level string) Option {
	return func(l *BaseLogger) {
		l.level = level
	}
}

func WithAddSource(add bool) Option {
	return func(l *BaseLogger) {
		l.addSource = add
	}
}

func WithName(name string) Option {
	return func(bl *BaseLogger) {
		bl.name = name
	}
}

func WithContext(ctx context.Context) Option {
	return func(bl *BaseLogger) {
		bl.ctx = ctx
	}
}

func WithLoggerType(loggerType string) Option {
	return func(bl *BaseLogger) {
		bl.loggerType = loggerType
	}
}

func WithLoggerTypeConsole() Option {
	return func(bl *BaseLogger) {
		bl.loggerType = LoggerTypeConsole
	}
}

func WithLoggerTypePretty() Option {
	return func(bl *BaseLogger) {
		bl.loggerType = LoggerTypePretty
	}
}

func WithLoggerTypeJSON() Option {
	return func(bl *BaseLogger) {
		bl.loggerType = LoggerTypeJSON
	}
}

// WithHandlerWrapper wraps the base slog handler before focus/name handlers.
func WithHandlerWrapper(wrapper func(slog.Handler) slog.Handler) Option {
	return func(bl *BaseLogger) {
		bl.handlerWrapper = wrapper
	}
}

// WithWriter sets the output writer for the underlying slog handler.
func WithWriter(writer io.Writer) Option {
	return func(bl *BaseLogger) {
		if writer == nil {
			return
		}
		bl.stdout = writer
	}
}

// WithRichErrorHandler sets a custom handler function to
// extract attributes from errors
func WithRichErrorHandler(handler RichErrorHandler) Option {
	return func(bl *BaseLogger) {
		bl.richErrHandler = handler
	}
}
