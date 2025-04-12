package glog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var DefaultLogLevel = Info

// BaseLogger implements both Logger and LoggerProvider interfaces
type BaseLogger struct {
	mu       sync.RWMutex
	logger   *slog.Logger
	root     *BaseLogger
	loggers  map[string]*BaseLogger
	opts     *slog.HandlerOptions
	ctx      context.Context
	focused  bool
	focusMap map[string]bool
	stdout   io.Writer

	level      string
	addSource  bool
	loggerType string
	name       string
}

func Argument[T any](key string, value T) any {
	return slog.Any(key, value)
}

func NewLogger(options ...Option) *BaseLogger {
	c := &BaseLogger{
		ctx:       context.Background(),
		level:     DefaultLogLevel,
		addSource: true,
		loggers:   map[string]*BaseLogger{},
		focusMap:  map[string]bool{},
		stdout:    os.Stdout,
	}

	for _, option := range options {
		option(c)
	}

	c.configureLogger()

	// TODO: refactor rename root to parent
	// TODO: refactor root should have not parent
	if c.root == nil {
		c.root = c
	}

	return c
}

// WithLevel sets the log level and returns the logger
func (c *BaseLogger) WithLevel(level string) *BaseLogger {
	c.level = level
	c.configureLogger()
	return c
}

// WithLevel sets the log level and returns the logger
func (c *BaseLogger) WithContext(ctx context.Context) Logger {
	newLogger := &BaseLogger{
		logger:     c.logger,
		root:       c.root,
		loggers:    c.loggers,
		opts:       c.opts,
		ctx:        ctx,
		name:       c.name,
		focusMap:   c.focusMap,
		level:      c.level,
		addSource:  c.addSource,
		loggerType: c.loggerType,
	}
	return newLogger
}

func (c *BaseLogger) WithLoggerType(loggerType string) Logger {
	c.loggerType = loggerType
	c.configureLogger()
	return c
}

func (c *BaseLogger) getRoot() *BaseLogger {
	if c.root == nil {
		return c
	}
	return c.root
}

func (c *BaseLogger) Focus(names ...string) {
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	root.focused = true
	root.focusMap = map[string]bool{}
	for _, name := range names {
		root.focusMap[name] = true
	}

	// TODO: Move to configureLogger
	for _, logger := range root.loggers {
		logger.configureLogger()
	}

	root.configureLogger()
}

func (c *BaseLogger) Unfocus() {
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	root.focused = false
	root.focusMap = map[string]bool{}

	for _, logger := range root.loggers {
		logger.configureLogger()
	}
	root.configureLogger()
}

func (c *BaseLogger) isFocused() bool {
	root := c.getRoot()
	root.mu.RLock()
	defer root.mu.RUnlock()

	if !root.focused {
		return true
	}

	return root.focusMap[c.name]
}

func (c *BaseLogger) GetLogger(name string) *BaseLogger {
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	if out, ok := c.root.loggers[name]; ok {
		return out
	}

	out := NewLogger()
	out.root = root
	out.name = name
	out.level = c.level
	out.addSource = c.addSource
	out.loggerType = c.loggerType

	out.configureLogger()

	c.root.loggers[name] = out

	return out
}

func (c *BaseLogger) Trace(msg string, args ...any) {
	c.logger.Log(c.ctx, LevelTrace, msg, args...)
}

func (c *BaseLogger) Debug(msg string, args ...any) {
	c.logger.Log(c.ctx, slog.LevelDebug, msg, args...)
}

func (c *BaseLogger) Info(msg string, args ...any) {
	c.logger.Log(c.ctx, slog.LevelInfo, msg, args...)
}

func (c *BaseLogger) Warn(msg string, args ...any) {
	c.logger.Log(c.ctx, slog.LevelWarn, msg, args...)
}

func (c *BaseLogger) Error(msg string, args ...any) {
	c.logger.Log(c.ctx, slog.LevelError, msg, args...)
}

func (c *BaseLogger) configureLogger() {
	c.opts = &slog.HandlerOptions{
		Level:     getLevel(c.level),
		AddSource: c.addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

			// Replace msg key with message string
			if a.Key == slog.TimeKey {
				a.Key = "ts"
				return a
			}

			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := CustomLevels[level]
				if !exists {
					levelLabel = level.String()
				}

				a.Value = slog.StringValue(strings.ToLower(levelLabel))
			}
			return a
		},
	}

	var handler slog.Handler

	switch c.loggerType {
	case LoggerTypeConsole:
		handler = slog.NewTextHandler(c.stdout, c.opts)
	case LoggerTypePretty:
		handler = NewColorConsoleHandler(c.stdout, c.opts)
	case LoggerTypeJSON:
		handler = slog.NewJSONHandler(c.stdout, c.opts)
	default:
		handler = slog.NewJSONHandler(c.stdout, c.opts)
	}

	handler = NewFocusFilterHandler(handler, c)

	if c.name != "" {
		handler = handler.WithAttrs([]slog.Attr{slog.String("logger", c.name)})
	}

	c.logger = slog.New(handler)
}

func NewFocusFilterHandler(handler slog.Handler, logger *BaseLogger) slog.Handler {
	return &FocusFilterHandler{
		handler: handler,
		logger:  logger,
	}
}

type FocusFilterHandler struct {
	handler slog.Handler
	logger  *BaseLogger
}

func (h *FocusFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if !h.handler.Enabled(ctx, level) {
		return false
	}
	return h.logger.isFocused()
}

func (h *FocusFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.logger.isFocused() {
		return nil
	}
	return h.handler.Handle(ctx, r)
}

func (h *FocusFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FocusFilterHandler{
		handler: h.handler.WithAttrs(attrs),
		logger:  h.logger,
	}
}

func (h *FocusFilterHandler) WithGroup(name string) slog.Handler {
	return &FocusFilterHandler{
		handler: h.handler.WithGroup(name),
		logger:  h.logger,
	}
}

func getLevel(l string) slog.Level {
	switch strings.ToUpper(l) {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "INFO":
		return slog.LevelInfo
	case "DEBUG":
		return slog.LevelDebug
	case "TRACE":
		return LevelTrace
	default:
		return slog.LevelInfo
	}
}
