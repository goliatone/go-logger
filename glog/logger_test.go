package glog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer, options ...Option) *BaseLogger {
	opts := append([]Option{func(bl *BaseLogger) {
		bl.stdout = buf
	}}, options...)
	return NewLogger(opts...)
}

func TestNewLogger(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf)

		assert.NotNil(t, logger.logger)
		assert.Equal(t, DefaultLogLevel, logger.level)
		assert.Equal(t, false, logger.addSource)
		assert.Equal(t, "", logger.loggerType)
		assert.Equal(t, logger, logger.root)
	})

	t.Run("with options", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := context.WithValue(context.Background(), "testkey", "testvalue")
		logger := newTestLogger(&buf,
			WithName("my-app"),
			WithLevel(Debug),
			WithLoggerTypePretty(),
			WithContext(ctx),
		)

		assert.Equal(t, "my-app", logger.name)
		assert.Equal(t, Debug, logger.level)
		assert.Equal(t, LoggerTypePretty, logger.loggerType)
		assert.Equal(t, "testvalue", logger.ctx.Value("testkey"))
	})
}

func TestGetLogger(t *testing.T) {
	var buf bytes.Buffer
	rootLogger := newTestLogger(&buf, WithName("root"), WithLevel(Info), WithLoggerTypePretty())

	child1 := rootLogger.GetLogger("child1")
	child1Logger, ok := child1.(*BaseLogger)
	require.True(t, ok)
	assert.Equal(t, "child1", child1Logger.name)
	assert.Equal(t, rootLogger, child1Logger.root)
	assert.Equal(t, rootLogger.level, child1Logger.level)
	assert.Equal(t, rootLogger.loggerType, child1Logger.loggerType)

	child1Again := rootLogger.GetLogger("child1")
	child1AgainLogger, ok := child1Again.(*BaseLogger)
	require.True(t, ok)
	assert.Same(t, child1Logger, child1AgainLogger)

	child1Logger.Info("hello from child")
	output := buf.String()
	assert.Contains(t, output, "[child1]")
	assert.Contains(t, output, "hello from child")
}

func TestLoggingLevels(t *testing.T) {
	testCases := []struct {
		name          string
		logLevel      string
		logFunc       func(l *BaseLogger, msg string, args ...any)
		expectedToLog bool
	}{
		{"Trace logs at TRACE level", Trace, (*BaseLogger).Trace, true},
		{"Debug logs at TRACE level", Trace, (*BaseLogger).Debug, true},
		{"Info logs at TRACE level", Trace, (*BaseLogger).Info, true},
		{"Debug logs at DEBUG level", Debug, (*BaseLogger).Debug, true},
		{"Info logs at DEBUG level", Debug, (*BaseLogger).Info, true},
		{"Trace logs NOT at DEBUG level", Debug, (*BaseLogger).Trace, false},
		{"Info logs at INFO level", Info, (*BaseLogger).Info, true},
		{"Warn logs at INFO level", Info, (*BaseLogger).Warn, true},
		{"Debug logs NOT at INFO level", Info, (*BaseLogger).Debug, false},
		{"Warn logs at WARN level", Warn, (*BaseLogger).Warn, true},
		{"Error logs at WARN level", Warn, (*BaseLogger).Error, true},
		{"Info logs NOT at WARN level", Warn, (*BaseLogger).Info, false},
		{"Error logs at ERROR level", Error, (*BaseLogger).Error, true},
		{"Warn logs NOT at ERROR level", Error, (*BaseLogger).Warn, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(&buf, WithLevel(tc.logLevel), WithLoggerTypeConsole())

			tc.logFunc(logger, "test message")

			output := buf.String()
			if tc.expectedToLog {
				assert.Contains(t, output, "test message")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestWithErrorLogging(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())

	logger.Error("something bad happened", wrappedErr)

	output := buf.String()
	assert.Contains(t, output, `"msg":"something bad happened"`)
	assert.Contains(t, output, `"error":"wrapped: base error"`)
	assert.Contains(t, output, `"root_error":"base error"`)
	assert.Contains(t, output, `"stack":`)
	assert.Contains(t, output, `glog.TestWithErrorLogging`)

	buf.Reset()
	logger.Error("something bad happened", "error", wrappedErr)

	output = buf.String()
	assert.Contains(t, output, `"msg":"something bad happened"`)
	assert.Contains(t, output, `"error":"wrapped: base error"`)
	assert.Contains(t, output, `"root_error":"base error"`)
	assert.Contains(t, output, `"stack":`)
	assert.Contains(t, output, `glog.TestWithErrorLogging`)
}

type uncomparableError []string

func (e uncomparableError) Error() string {
	return strings.Join(e, ": ")
}

type uncomparableWrappedError []string

func (e uncomparableWrappedError) Error() string {
	return strings.Join(e, ": ")
}

func (e uncomparableWrappedError) Unwrap() error {
	return errors.New("uncomparable root")
}

type cyclicError struct{}

func (cyclicError) Error() string   { return "cyclic error" }
func (e cyclicError) Unwrap() error { return e }

func TestErrorLoggingAcceptsUncomparableErrors(t *testing.T) {
	err := uncomparableError{"readiness failed", "archive route index unavailable"}

	for _, tt := range []struct {
		name string
		args []any
	}{
		{name: "bare error", args: []any{err}},
		{name: "keyed error", args: []any{"error", err}},
		{name: "slog attribute", args: []any{slog.Any("error", err)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newTestLogger(&buf, WithLoggerTypeJSON())

			assert.NotPanics(t, func() {
				logger.Error("optional capability refresh failed", tt.args...)
			})

			output := buf.String()
			assert.Contains(t, output, `"msg":"optional capability refresh failed"`)
			assert.Contains(t, output, `"error":"readiness failed: archive route index unavailable"`)
			assert.NotContains(t, output, `"root_error"`)
		})
	}
}

func TestErrorLoggingTraversesUncomparableAndMultiErrorCauses(t *testing.T) {
	t.Run("uncomparable linear wrapper", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		logger.Error("linear failure", uncomparableWrappedError{"wrapped"})

		output := buf.String()
		assert.Contains(t, output, `"root_error":"uncomparable root"`)
		assert.NotContains(t, output, `"error_causes"`)
	})

	t.Run("branching multi error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		first := errors.New("first cause")
		second := errors.New("second cause")
		logger.Error("multi failure", errors.Join(first, fmt.Errorf("wrapped: %w", second)))

		output := buf.String()
		assert.Contains(t, output, `"error_causes":["first cause","second cause"]`)
		assert.NotContains(t, output, `"root_error"`)
	})

	t.Run("cyclic chain is bounded", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		logger.Error("cyclic failure", cyclicError{})

		output := buf.String()
		assert.Contains(t, output, `"root_error":"cyclic error"`)
		assert.Contains(t, output, `"error_causes_truncated":true`)
	})

	t.Run("cyclic branch does not hide later siblings", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		logger.Error("cyclic multi failure", errors.Join(cyclicError{}, errors.New("healthy sibling")))

		output := buf.String()
		assert.Contains(t, output, `"error_causes":["cyclic error","healthy sibling"]`)
		assert.Contains(t, output, `"error_causes_truncated":true`)
	})

	t.Run("wide branch preserves nested leaf order", func(t *testing.T) {
		children := make([]error, maxErrorCauseChildren)
		children[0] = errors.Join(errors.New("nested-a"), errors.New("nested-b"))
		for i := 1; i < len(children); i++ {
			children[i] = fmt.Errorf("sibling-%02d", i)
		}

		inspection := inspectErrorCauses(errors.Join(children...))
		expected := []string{"nested-a", "nested-b"}
		for i := 1; len(expected) < maxErrorCauseLeaves; i++ {
			expected = append(expected, fmt.Sprintf("sibling-%02d", i))
		}
		assert.Equal(t, expected, inspection.leaves)
		assert.True(t, inspection.truncated)
	})

	t.Run("cyclic branch does not flatten a wrapped sibling", func(t *testing.T) {
		inspection := inspectErrorCauses(errors.Join(
			cyclicError{},
			fmt.Errorf("wrapped sibling: %w", errors.New("healthy leaf")),
		))

		assert.Equal(t, []string{"cyclic error", "healthy leaf"}, inspection.leaves)
		assert.True(t, inspection.truncated)
	})

	t.Run("leaf count is bounded", func(t *testing.T) {
		causes := make([]error, maxErrorCauseLeaves+2)
		for i := range causes {
			causes[i] = fmt.Errorf("cause-%02d", i)
		}
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		logger.Error("large multi failure", errors.Join(causes...))

		output := buf.String()
		assert.Contains(t, output, `"error_causes_truncated":true`)
		assert.Contains(t, output, `"cause-15"`)
		assert.NotContains(t, output, `"cause-16"`)
	})

	t.Run("leaf text is bounded", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newTestLogger(&buf, WithLoggerTypeJSON())
		logger.Error("large cause", fmt.Errorf("wrapper: %w", errors.New(strings.Repeat("x", maxErrorCauseTextBytes+10))))

		output := buf.String()
		assert.Contains(t, output, `"error_causes_truncated":true`)
		assert.Contains(t, output, `"root_error":"`)
	})
}

type customError struct {
	msg    string
	code   int
	status string
}

func (e customError) Error() string  { return e.msg }
func (e customError) Code() int      { return e.code }
func (e customError) Status() string { return e.status }

func TestWithRichErrorHandler(t *testing.T) {

	err := customError{msg: "db error", code: 5001, status: "INTERNAL_SERVER_ERROR"}

	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())
	logger.Error("request failed", err)

	output := buf.String()
	assert.Contains(t, output, `"error_code":5001`)
	assert.Contains(t, output, `"status_code":"INTERNAL_SERVER_ERROR"`)

	buf.Reset()
	customHandler := func(err error) []slog.Attr {
		if ce, ok := err.(customError); ok {
			return []slog.Attr{slog.String("custom_key", "custom_value"), slog.Int("error_code", ce.code)}
		}
		return nil
	}
	loggerWithCustomHandler := newTestLogger(&buf, WithLoggerTypeJSON(), WithRichErrorHandler(customHandler))
	loggerWithCustomHandler.Error("request failed again", err)

	output = buf.String()
	assert.Contains(t, output, `"custom_key":"custom_value"`)
	assert.Contains(t, output, `"error_code":5001`)
	assert.NotContains(t, output, "status_code")
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())

	loggerWithAttrs := logger.With("user_id", 123, "request_id", "abc-123")
	loggerWithAttrs.Info("user logged in")

	output := buf.String()
	assert.Contains(t, output, `"user_id":123`)
	assert.Contains(t, output, `"request_id":"abc-123"`)
	assert.Contains(t, output, `"msg":"user logged in"`)

	buf.Reset()
	logger.Info("system check")
	output = buf.String()
	assert.NotContains(t, output, "user_id")
	assert.Contains(t, output, `"msg":"system check"`)
}

func TestWithFieldsDeterministicOrdering(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())

	fields := map[string]any{
		"zeta":   3,
		"alpha":  1,
		"middle": 2,
	}

	logger.WithFields(fields).Info("with fields")

	output := buf.String()
	require.Contains(t, output, `"alpha":1`)
	require.Contains(t, output, `"middle":2`)
	require.Contains(t, output, `"zeta":3`)

	idxAlpha := strings.Index(output, `"alpha":1`)
	idxMiddle := strings.Index(output, `"middle":2`)
	idxZeta := strings.Index(output, `"zeta":3`)
	assert.True(t, idxAlpha < idxMiddle && idxMiddle < idxZeta, "fields should be sorted alphabetically")

	buf.Reset()
	shuffled := map[string]any{
		"middle": 2,
		"zeta":   3,
		"alpha":  1,
	}
	logger.WithFields(shuffled).Info("with fields again")

	output = buf.String()
	require.Contains(t, output, `"alpha":1`)
	require.Contains(t, output, `"middle":2`)
	require.Contains(t, output, `"zeta":3`)

	idxAlpha = strings.Index(output, `"alpha":1`)
	idxMiddle = strings.Index(output, `"middle":2`)
	idxZeta = strings.Index(output, `"zeta":3`)
	assert.True(t, idxAlpha < idxMiddle && idxMiddle < idxZeta, "fields should retain sorted order across calls")
}

func TestWithFieldsEmptyMap(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())

	out := logger.WithFields(nil)
	require.IsType(t, (*BaseLogger)(nil), out)
	assert.Same(t, logger, out.(*BaseLogger))

	empty := map[string]any{}
	out = logger.WithFields(empty)
	require.IsType(t, (*BaseLogger)(nil), out)
	assert.Same(t, logger, out.(*BaseLogger))
}

func TestWithFieldsComposition(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON())

	logger.With("user_id", 123).
		WithFields(map[string]any{"request_id": "abc-123"}).
		Info("user action")

	output := buf.String()
	assert.Contains(t, output, `"user_id":123`)
	assert.Contains(t, output, `"request_id":"abc-123"`)
	assert.Contains(t, output, `"msg":"user action"`)
}

func TestArgsHelperExpansion(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeJSON(), WithLevel(Debug))

	logger.Debug("request processed", Args(
		"method", "POST",
		"status", 200,
	))

	output := buf.String()
	assert.Contains(t, output, `"method":"POST"`)
	assert.Contains(t, output, `"status":200`)
}

func TestFocusAndUnfocus(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(&buf, WithLoggerTypePretty(), WithLevel(Debug))
	apiLogger := root.GetLogger("api")
	dbLogger := root.GetLogger("db")
	cacheLogger := root.GetLogger("cache")

	root.Focus("api", "db")

	apiLogger.Info("api call")
	assert.Contains(t, buf.String(), "[api]")
	buf.Reset()

	dbLogger.Info("db query")
	assert.Contains(t, buf.String(), "[db]")
	buf.Reset()

	cacheLogger.Info("cache hit")
	assert.Empty(t, buf.String())
	buf.Reset()

	root.Unfocus()

	apiLogger.Info("api call after unfocus")
	assert.Contains(t, buf.String(), "[api]")
	buf.Reset()

	dbLogger.Info("db query after unfocus")
	assert.Contains(t, buf.String(), "[db]")
	buf.Reset()

	cacheLogger.Info("cache hit after unfocus")
	assert.Contains(t, buf.String(), "[cache]")
	buf.Reset()
}

func TestFatal(t *testing.T) {
	var exitCode int
	var exitCalled bool
	var mu sync.Mutex

	exitFunc := func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
		exitCalled = true
	}

	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeConsole(), WithExitFunc(exitFunc))

	logger.Fatal("critical failure")

	mu.Lock()
	defer mu.Unlock()

	assert.True(t, exitCalled, "exit func should have been called")
	assert.Equal(t, 1, exitCode, "default exit code should be 1")
	assert.Contains(t, buf.String(), "critical failure")
}

func TestFatalBehaviorLogOnly(t *testing.T) {
	var exitCalled bool

	var buf bytes.Buffer
	logger := newTestLogger(&buf,
		WithLoggerTypeJSON(),
		WithFatalBehavior(FatalBehaviorLogOnly),
		WithExitFunc(func(int) {
			exitCalled = true
		}),
	)

	logger.Fatal("fatal but no exit")

	assert.False(t, exitCalled, "exit func should not be called for log-only fatal behavior")
	assert.Contains(t, buf.String(), `"msg":"fatal but no exit"`)
}

func TestFatalBehaviorPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf,
		WithLoggerTypeJSON(),
		WithFatalBehavior(FatalBehaviorPanic),
	)

	assert.PanicsWithValue(t, "panic now", func() {
		logger.Fatal("panic now")
	})
}

func TestFatalLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf,
		WithLoggerTypeJSON(),
		WithLevel(Fatal),
		WithFatalBehavior(FatalBehaviorLogOnly),
	)

	logger.Error("should be hidden", errors.New("boom"))
	assert.Empty(t, buf.String())

	logger.Fatal("should be shown")
	output := buf.String()
	assert.Contains(t, output, `"msg":"should be shown"`)
	assert.Contains(t, output, `"level":"fatal"`)
}

func TestFindError(t *testing.T) {
	errSample := errors.New("i am an error")

	testCases := []struct {
		name          string
		args          []any
		expectedErr   error
		expectedRem   []any
		shouldMatch   bool // use for require.Equal because order matters
		shouldContain bool // use for assert.ElementsMatch because order does not matter
	}{
		{
			name:        "no error",
			args:        []any{"key", "value", 123},
			expectedErr: nil,
			expectedRem: []any{"key", "value", 123},
			shouldMatch: true,
		},
		{
			name:        "error is only arg",
			args:        []any{errSample},
			expectedErr: errSample,
			expectedRem: []any{},
			shouldMatch: true,
		},
		{
			name:          "error is among other args",
			args:          []any{"key", "value", errSample, "other", 1},
			expectedErr:   errSample,
			expectedRem:   []any{"key", "value", "other", 1},
			shouldContain: true,
		},
		{
			name:        "error is in key-value pair",
			args:        []any{"request_id", "123", "error", errSample},
			expectedErr: errSample,
			expectedRem: []any{"request_id", "123"},
			shouldMatch: true,
		},
		{
			name:        "error is in slog.Attr",
			args:        []any{slog.Any("error", errSample), "request_id", "123"},
			expectedErr: errSample,
			expectedRem: []any{"request_id", "123"},
			shouldMatch: true,
		},
		{
			name:        "first error is found",
			args:        []any{errors.New("first"), errors.New("second")},
			expectedErr: errors.New("first"),
			expectedRem: []any{errors.New("second")},
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err, rem := findError(tc.args)
			if tc.expectedErr == nil {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedErr.Error())
			}
			if tc.shouldMatch {
				require.Equal(t, tc.expectedRem, rem)
			}
			if tc.shouldContain {
				assert.ElementsMatch(t, tc.expectedRem, rem)
			}
		})
	}
}

func TestWithLevelPropagatesToChildren(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(&buf, WithLevel(Info), WithLoggerTypeJSON())
	child := root.GetLogger("child").(*BaseLogger)

	root.WithLevel(Debug)

	assert.Equal(t, Debug, root.level)
	assert.Equal(t, Debug, child.level)
}

func TestWithLoggerTypePropagatesToChildren(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(&buf, WithLoggerTypeJSON())
	child := root.GetLogger("child").(*BaseLogger)

	root.WithLoggerType(LoggerTypePretty)

	assert.Equal(t, LoggerTypePretty, root.loggerType)
	assert.Equal(t, LoggerTypePretty, child.loggerType)
}

type spyHandler struct {
	handler slog.Handler
	called  *bool
}

func (s *spyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return s.handler.Enabled(ctx, level)
}

func (s *spyHandler) Handle(ctx context.Context, r slog.Record) error {
	if s.called != nil {
		*s.called = true
	}
	return s.handler.Handle(ctx, r)
}

func (s *spyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &spyHandler{handler: s.handler.WithAttrs(attrs), called: s.called}
}

func (s *spyHandler) WithGroup(name string) slog.Handler {
	return &spyHandler{handler: s.handler.WithGroup(name), called: s.called}
}

func TestWithHandlerWrapper(t *testing.T) {
	var buf bytes.Buffer
	var wrappedType string
	handlerCalled := false

	logger := NewLogger(
		WithLoggerTypeConsole(),
		WithWriter(&buf),
		WithHandlerWrapper(func(h slog.Handler) slog.Handler {
			wrappedType = fmt.Sprintf("%T", h)
			return &spyHandler{handler: h, called: &handlerCalled}
		}),
	)

	logger.Info("wrapped handler test")

	assert.True(t, handlerCalled, "wrapper handler should be invoked")
	assert.Equal(t, "*slog.TextHandler", wrappedType)
	assert.NotEmpty(t, buf.String())
	assert.Contains(t, buf.String(), "wrapped handler test")
}

type capturedLogRecord struct {
	record slog.Record
	attrs  []slog.Attr
	groups []string
}

type recordCaptureSink struct {
	mu      sync.Mutex
	records []capturedLogRecord
}

type recordCaptureHandler struct {
	sink   *recordCaptureSink
	next   slog.Handler
	attrs  []slog.Attr
	groups []string
}

func (h *recordCaptureHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next == nil || h.next.Enabled(ctx, level)
}

func (h *recordCaptureHandler) Handle(ctx context.Context, record slog.Record) error {
	h.sink.mu.Lock()
	h.sink.records = append(h.sink.records, capturedLogRecord{
		record: record.Clone(),
		attrs:  append([]slog.Attr(nil), h.attrs...),
		groups: append([]string(nil), h.groups...),
	})
	h.sink.mu.Unlock()
	if h.next == nil {
		return nil
	}
	return h.next.Handle(ctx, record)
}

func (h *recordCaptureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	if h.next != nil {
		clone.next = h.next.WithAttrs(attrs)
	}
	return &clone
}

func (h *recordCaptureHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string(nil), h.groups...), name)
	if h.next != nil {
		clone.next = h.next.WithGroup(name)
	}
	return &clone
}

func (s *recordCaptureSink) snapshot(t *testing.T) []capturedLogRecord {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]capturedLogRecord, len(s.records))
	copy(out, s.records)
	return out
}

func recordAttrs(record slog.Record) map[string]any {
	attrs := map[string]any{}
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})
	return attrs
}

func recordFunction(record slog.Record) string {
	if record.PC == 0 {
		return ""
	}
	frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
	return frame.Function
}

//go:noinline
func logInfoThroughAdapter(logger *BaseLogger) {
	logger.Info("through adapter")
}

//go:noinline
func logErrorThroughAdapter(logger *BaseLogger, err error) {
	logger.Error("adapter failed", err)
}

func TestWithCallerSkipNormalizationAndPropagation(t *testing.T) {
	var buf bytes.Buffer

	negative := newTestLogger(&buf, WithCallerSkip(-2))
	assert.Equal(t, 0, negative.callerSkip)

	logger := newTestLogger(&buf, WithCallerSkip(maxAdditionalCallerSkip+100))
	assert.Equal(t, maxAdditionalCallerSkip, logger.callerSkip)

	logger = newTestLogger(&buf, WithCallerSkip(2))
	assert.Equal(t, 2, logger.WithContext(context.Background()).(*BaseLogger).callerSkip)
	assert.Equal(t, 2, logger.GetLogger("child").(*BaseLogger).callerSkip)
	assert.Equal(t, 2, logger.With("component", "worker").callerSkip)
	assert.Equal(t, 2, logger.WithFields(map[string]any{"request_id": "req-1"}).(*BaseLogger).callerSkip)
}

func TestWithCallerSkipResolvesAdapterCallerAndStack(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithCallerSkip(1),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	logInfoThroughAdapter(logger)
	logErrorThroughAdapter(logger, fmt.Errorf("wrapped: %w", errors.New("root")))

	records := sink.snapshot(t)
	require.Len(t, records, 2)
	assert.Contains(t, recordFunction(records[0].record), "TestWithCallerSkipResolvesAdapterCallerAndStack")
	assert.Contains(t, recordFunction(records[1].record), "TestWithCallerSkipResolvesAdapterCallerAndStack")

	attrs := recordAttrs(records[1].record)
	require.Contains(t, attrs, "stack")
	stack, ok := attrs["stack"].(string)
	require.True(t, ok)
	assert.Contains(t, stack, "TestWithCallerSkipResolvesAdapterCallerAndStack")
	assert.NotContains(t, stack, "logErrorThroughAdapter")
}

func TestDefaultCallerRemainsApplicationCallSite(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithFatalBehavior(FatalBehaviorLogOnly),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	logger.Info("direct info")
	logger.Error("direct error", errors.New("boom"))
	logger.Fatal("direct fatal", errors.New("fatal boom"))

	records := sink.snapshot(t)
	require.Len(t, records, 3)
	for _, captured := range records {
		assert.Contains(t, recordFunction(captured.record), "TestDefaultCallerRemainsApplicationCallSite")
	}
	assert.Contains(t, recordAttrs(records[1].record)["stack"], "TestDefaultCallerRemainsApplicationCallSite")
	assert.Contains(t, recordAttrs(records[2].record)["stack"], "TestDefaultCallerRemainsApplicationCallSite")
}

func TestHandlerWrapperObservesCanonicalEnrichedRecordOnce(t *testing.T) {
	var buf bytes.Buffer
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithLoggerTypeJSON(),
		WithWriter(&buf),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{
				sink:  sink,
				next:  next,
				attrs: []slog.Attr{slog.String("delegate", fmt.Sprintf("%T", next))},
			}
		}),
	)

	logger.GetLogger("api").(*BaseLogger).
		With("request_id", "req-1").
		Error("request failed", fmt.Errorf("outer: %w", errors.New("root cause")))

	records := sink.snapshot(t)
	require.Len(t, records, 1)
	require.NotZero(t, records[0].record.PC)
	attrs := recordAttrs(records[0].record)
	assert.Equal(t, "root cause", attrs["root_error"])
	assert.EqualError(t, attrs["error"].(error), "outer: root cause")
	assert.NotEmpty(t, attrs["stack"])
	assert.Equal(t, 1, strings.Count(buf.String(), `"msg":"request failed"`))

	boundAttrs := map[string]any{}
	for _, attr := range records[0].attrs {
		boundAttrs[attr.Key] = attr.Value.Any()
	}
	assert.Equal(t, "api", boundAttrs["logger"])
	assert.Equal(t, "req-1", boundAttrs["request_id"])
}

func TestHandlerWrapperPreservesAttrsAndGroups(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithWriter(io.Discard),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	handler := logger.logger.Handler().
		WithGroup("request").
		WithAttrs([]slog.Attr{slog.String("request_id", "req-1")})
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "grouped", 0)
	require.NoError(t, handler.Handle(context.Background(), record))

	records := sink.snapshot(t)
	require.Len(t, records, 1)
	assert.Equal(t, []string{"request"}, records[0].groups)
	require.Len(t, records[0].attrs, 1)
	assert.Equal(t, "request_id", records[0].attrs[0].Key)
	assert.Equal(t, "req-1", records[0].attrs[0].Value.String())
}

func TestCallerSkipAndHandlerCaptureAreConcurrentSafe(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithCallerSkip(1),
		WithWriter(io.Discard),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	const callers = 24
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := range callers {
		go func() {
			defer wg.Done()
			logInfoThroughAdapter(logger.With("worker", i))
		}()
	}
	wg.Wait()

	assert.Len(t, sink.snapshot(t), callers)
}

type orderedHandler struct {
	next  slog.Handler
	name  string
	order *[]string
}

func (h *orderedHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *orderedHandler) Handle(ctx context.Context, record slog.Record) error {
	*h.order = append(*h.order, h.name)
	return h.next.Handle(ctx, record)
}

func (h *orderedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &orderedHandler{next: h.next.WithAttrs(attrs), name: h.name, order: h.order}
}

func (h *orderedHandler) WithGroup(name string) slog.Handler {
	return &orderedHandler{next: h.next.WithGroup(name), name: h.name, order: h.order}
}

func TestWithHandlerWrapperCompositionOrder(t *testing.T) {
	var buf bytes.Buffer
	order := []string{}
	logger := NewLogger(
		WithWriter(&buf),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &orderedHandler{next: next, name: "first", order: &order}
		}),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &orderedHandler{next: next, name: "second", order: &order}
		}),
	)

	logger.With("request_id", "req-1").Info("composed")

	assert.Equal(t, []string{"second", "first"}, order)
	assert.Contains(t, buf.String(), `"request_id":"req-1"`)
}

func TestWithWriter(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(
		WithLoggerTypeJSON(),
		WithWriter(&buf),
	)

	logger.Info("writer override test")

	assert.Contains(t, buf.String(), `"msg":"writer override test"`)
}
