package glog

import "context"

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

// WithRichErrorHandler sets a custom handler function to
// extract attributes from errors
func WithRichErrorHandler(handler RichErrorHandler) Option {
	return func(bl *BaseLogger) {
		bl.richErrHandler = handler
	}
}
