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

// WithCallerSkip adds known adapter frames to the caller depth used for
// slog.Record.PC and error stack capture. Values below zero are treated as
// zero and excessive values are bounded. Use this only for stable wrapper
// layers owned by the application; it is not a general stack-hiding feature.
func WithCallerSkip(skip int) Option {
	return func(l *BaseLogger) {
		l.callerSkip = normalizeAdditionalCallerSkip(skip)
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
// Multiple options compose in declaration order: the first wrapper is closest
// to the base handler and each later wrapper becomes the new outer wrapper.
func WithHandlerWrapper(wrapper func(slog.Handler) slog.Handler) Option {
	return func(bl *BaseLogger) {
		if wrapper == nil {
			return
		}
		if bl.handlerWrapper == nil {
			bl.handlerWrapper = wrapper
			return
		}

		inner := bl.handlerWrapper
		bl.handlerWrapper = func(handler slog.Handler) slog.Handler {
			return wrapper(inner(handler))
		}
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

// WithExitFunc overrides the default exit behavior for Fatal.
func WithExitFunc(exit func(int)) Option {
	return func(bl *BaseLogger) {
		if exit == nil {
			return
		}
		bl.exitFunc = exit
	}
}

// WithFatalBehavior configures what Fatal does after logging.
func WithFatalBehavior(behavior FatalBehavior) Option {
	return func(bl *BaseLogger) {
		bl.fatalBehavior = behavior
	}
}
