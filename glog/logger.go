package glog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var DefaultLogLevel = Info

// osExit is a mockable exit function.
var osExit = os.Exit

type RichErrorHandler func(err error) []slog.Attr

// BaseLogger implements both Logger and LoggerProvider interfaces
type BaseLogger struct {
	mu             *sync.RWMutex
	logger         *slog.Logger
	root           *BaseLogger
	loggers        map[string]*BaseLogger
	opts           *slog.HandlerOptions
	ctx            context.Context
	focused        bool
	focusMap       map[string]bool
	stdout         io.Writer
	handlerWrapper func(slog.Handler) slog.Handler
	exitFunc       func(int)
	fatalBehavior  FatalBehavior

	level      string
	addSource  bool
	loggerType string
	name       string

	richErrHandler RichErrorHandler
}

var _ FieldsLogger = (*BaseLogger)(nil)
var _ Logger = (*BaseLogger)(nil)
var _ LoggerProvider = (*BaseLogger)(nil)

type ArgsList []any

func Arg(key string, value any) any {
	return slog.Any(key, value)
}

func Args(args ...any) ArgsList {
	return ArgsList(argsToAttrSlice(args))
}

func NewLogger(options ...Option) *BaseLogger {
	c := &BaseLogger{
		mu:             &sync.RWMutex{},
		ctx:            context.Background(),
		level:          DefaultLogLevel,
		addSource:      false,
		loggers:        map[string]*BaseLogger{},
		focusMap:       map[string]bool{},
		stdout:         os.Stdout,
		richErrHandler: defaultErrHandler,
		fatalBehavior:  FatalBehaviorExit,
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
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	c.level = level
	c.configureLogger()

	if c == root {
		for _, logger := range root.loggers {
			logger.level = level
			logger.configureLogger()
		}
	}
	return c
}

// WithLevel sets the log level and returns the logger
func (c *BaseLogger) WithContext(ctx context.Context) Logger {
	newLogger := &BaseLogger{
		logger:         c.logger,
		root:           c.root,
		loggers:        c.loggers,
		opts:           c.opts,
		ctx:            ctx,
		name:           c.name,
		focusMap:       c.focusMap,
		level:          c.level,
		addSource:      c.addSource,
		loggerType:     c.loggerType,
		stdout:         c.stdout,
		handlerWrapper: c.handlerWrapper,
		richErrHandler: c.richErrHandler,
		exitFunc:       c.exitFunc,
		fatalBehavior:  c.fatalBehavior,
		mu:             c.mu, // we share the mutex pointer
	}
	return newLogger
}

func (c *BaseLogger) WithLoggerType(loggerType string) Logger {
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	c.loggerType = loggerType
	c.configureLogger()

	if c == root {
		for _, logger := range root.loggers {
			logger.loggerType = loggerType
			logger.configureLogger()
		}
	}
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

func (c *BaseLogger) GetLogger(name string) Logger {
	root := c.getRoot()
	root.mu.Lock()
	defer root.mu.Unlock()

	if out, ok := c.root.loggers[name]; ok {
		return out
	}

	out := &BaseLogger{
		ctx:            c.ctx,
		stdout:         c.stdout,
		handlerWrapper: c.handlerWrapper,
		richErrHandler: c.richErrHandler,
		loggers:        make(map[string]*BaseLogger),
		focusMap:       make(map[string]bool),
		root:           root,
		name:           name,
		level:          c.level,
		addSource:      c.addSource,
		loggerType:     c.loggerType,
		exitFunc:       c.exitFunc,
		fatalBehavior:  c.fatalBehavior,
		mu:             root.mu, // we share the root mutex
	}

	out.configureLogger()

	c.root.loggers[name] = out

	return out
}

// With returns a Logger that includes the given attributes
// in each subsequent log output.
func (c *BaseLogger) With(args ...any) *BaseLogger {
	if len(args) == 0 {
		return c
	}

	c2 := *c
	c2.logger = c.logger.With(argsToAttrSlice(args)...)
	return &c2
}

// WithFields returns a logger that includes the given structured fields.
func (c *BaseLogger) WithFields(fields map[string]any) Logger {
	if len(fields) == 0 {
		return c
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	args := make([]any, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, k, fields[k])
	}

	return c.With(args...)
}

func (c *BaseLogger) Trace(msg string, args ...any) {
	c.log(c.ctx, LevelTrace, msg, args...)
}

func (c *BaseLogger) Debug(msg string, args ...any) {
	c.log(c.ctx, slog.LevelDebug, msg, args...)
}

func (c *BaseLogger) Info(msg string, args ...any) {
	c.log(c.ctx, slog.LevelInfo, msg, args...)
}

func (c *BaseLogger) Warn(msg string, args ...any) {
	c.log(c.ctx, slog.LevelWarn, msg, args...)
}

func (c *BaseLogger) Error(msg string, args ...any) {
	c.errorWithSkip(slog.LevelError, msg, defaultSkipFrames, args...)
}

func (c *BaseLogger) Fatal(msg string, args ...any) {
	err := c.errorWithSkip(LevelFatal, msg, defaultSkipFrames, args...)

	switch c.fatalBehavior {
	case FatalBehaviorLogOnly:
		return
	case FatalBehaviorPanic:
		if err != nil {
			panic(err)
		}
		panic(msg)
	default:
		code := 1
		if err != nil {
			if ce, ok := err.(coder); ok {
				code = ce.Code()
			}
		}

		// NOTE: might need to come up with a way to flush any async logs, maybe
		exitFunc := c.exitFunc
		if exitFunc == nil {
			exitFunc = osExit
		}
		exitFunc(code)
	}
}

const (
	defaultSkipFrames = 4
)

// log is the low-level logging method that all other log methods call.
// It's responsible for creating the slog.Record with the correct caller PC
// and passing it to the handler. This bypasses slog.Logger.Log to avoid
// capturing the call site within this package
func (c *BaseLogger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	c.logWithSkip(ctx, level, msg, defaultSkipFrames, args...)
}

func (c *BaseLogger) logWithSkip(ctx context.Context, level slog.Level, msg string, skip int, args ...any) {
	normalized := normalizeArgs(args)
	c.logWithSkipNormalized(ctx, level, msg, skip, normalized)
}

func (c *BaseLogger) logWithSkipNormalized(ctx context.Context, level slog.Level, msg string, skip int, args []any) {
	if !c.logger.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	_ = c.logger.Handler().Handle(ctx, r)
}

func (c *BaseLogger) errorWithSkip(level slog.Level, msg string, skip int, args ...any) error {
	normalized := normalizeArgs(args)
	err, nargs := findError(normalized)
	if err == nil {
		c.logWithSkipNormalized(c.ctx, level, msg, skip, nargs)
		return nil
	}

	dargs := nargs

	if c.richErrHandler != nil {
		if richAttrs := c.richErrHandler(err); richAttrs != nil {
			for _, attr := range richAttrs {
				dargs = append(dargs, attr)
			}
		}
	}

	root := err
	for {
		unwrapped := errors.Unwrap(root)
		if unwrapped == nil {
			break
		}
		root = unwrapped
	}

	if root != err {
		dargs = append(dargs, slog.Any("root_error", root))
	}

	dargs = append(dargs, slog.Any("error", err))

	stack := getStackTrace(skip)

	dargs = append(dargs, slog.Any("stack", stack))

	c.logWithSkipNormalized(c.ctx, level, msg, skip, dargs)
	return err
}

func findError(args []any) (errFound error, remaining []any) {
	remaining = make([]any, 0, len(args))

	for i := 0; i < len(args); i++ {
		if key, ok := args[i].(string); ok && key == "error" && i+1 < len(args) {
			if errVal, ok := args[i+1].(error); ok && errVal != nil {
				if errFound == nil {
					errFound = errVal
					i++
					continue
				}
				remaining = append(remaining, args[i], args[i+1])
				i++
				continue
			}
		}

		if attr, ok := args[i].(slog.Attr); ok && attr.Key == "error" {
			if errVal, ok := attr.Value.Any().(error); ok && errVal != nil {
				if errFound == nil {
					errFound = errVal
					continue
				}
			}
			remaining = append(remaining, attr)
			continue
		}

		if e, ok := args[i].(error); ok && e != nil && errFound == nil {
			errFound = e
			continue
		}
		remaining = append(remaining, args[i])
	}
	return errFound, remaining
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

	if c.handlerWrapper != nil {
		handler = c.handlerWrapper(handler)
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
	case "FATAL":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

func getStackTrace(skip int) string {
	const depth = 32
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}

func defaultErrHandler(err error) []slog.Attr {
	var attrs []slog.Attr
	if ce, ok := err.(coder); ok {
		attrs = append(attrs, slog.Any("error_code", ce.Code()))
	}

	if ce, ok := err.(statuser); ok {
		attrs = append(attrs, slog.Any("status_code", ce.Status()))
	}

	if len(attrs) == 0 {
		return nil
	}

	return attrs
}
