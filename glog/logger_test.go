package glog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"

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
		assert.Equal(t, true, logger.addSource)
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
	assert.Equal(t, "child1", child1.name)
	assert.Equal(t, rootLogger, child1.root)
	assert.Equal(t, rootLogger.level, child1.level)
	assert.Equal(t, rootLogger.loggerType, child1.loggerType)

	child1Again := rootLogger.GetLogger("child1")
	assert.Same(t, child1, child1Again)

	child1.Info("hello from child")
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

	originalExit := osExit
	osExit = func(code int) {
		mu.Lock()
		defer mu.Unlock()
		exitCode = code
		exitCalled = true
	}
	t.Cleanup(func() {
		osExit = originalExit
	})

	var buf bytes.Buffer
	logger := newTestLogger(&buf, WithLoggerTypeConsole())

	logger.Fatal("critical failure")

	mu.Lock()
	defer mu.Unlock()

	assert.True(t, exitCalled, "os.Exit should have been called")
	assert.Equal(t, 1, exitCode, "default exit code should be 1")
	assert.Contains(t, buf.String(), "critical failure")
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
			expectedErr: nil,
			expectedRem: []any{"request_id", "123", "error", errSample},
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
