package compat

import (
	"context"
	"fmt"
	"testing"

	"github.com/goliatone/go-logger/glog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type logCall struct {
	level string
	msg   string
}

type recordingFormattedLogger struct {
	calls []logCall
}

func (l *recordingFormattedLogger) Debugf(format string, args ...any) {
	l.calls = append(l.calls, logCall{level: "debugf", msg: fmt.Sprintf(format, args...)})
}

func (l *recordingFormattedLogger) Infof(format string, args ...any) {
	l.calls = append(l.calls, logCall{level: "infof", msg: fmt.Sprintf(format, args...)})
}

func (l *recordingFormattedLogger) Warnf(format string, args ...any) {
	l.calls = append(l.calls, logCall{level: "warnf", msg: fmt.Sprintf(format, args...)})
}

func (l *recordingFormattedLogger) Errorf(format string, args ...any) {
	l.calls = append(l.calls, logCall{level: "errorf", msg: fmt.Sprintf(format, args...)})
}

type recordingLogger struct {
	calls []logCall
}

func (l *recordingLogger) Trace(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "trace", msg: msg})
}

func (l *recordingLogger) Debug(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "debug", msg: msg})
}

func (l *recordingLogger) Info(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "info", msg: msg})
}

func (l *recordingLogger) Warn(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "warn", msg: msg})
}

func (l *recordingLogger) Error(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "error", msg: msg})
}

func (l *recordingLogger) Fatal(msg string, args ...any) {
	l.calls = append(l.calls, logCall{level: "fatal", msg: msg})
}

func (l *recordingLogger) WithContext(ctx context.Context) glog.Logger {
	return l
}

func TestFromFormattedNilSafe(t *testing.T) {
	logger := FromFormatted(nil)
	require.NotNil(t, logger)
	assert.NotPanics(t, func() {
		logger.Info("safe")
		logger.WithContext(context.Background()).Debug("safe")
	})
}

func TestFromFormattedLevelMapping(t *testing.T) {
	formatted := &recordingFormattedLogger{}
	logger := FromFormatted(formatted)

	logger.Trace("trace %d", 1)
	logger.Debug("debug %d", 2)
	logger.Info("info %d", 3)
	logger.Warn("warn %d", 4)
	logger.Error("error %d", 5)
	logger.Fatal("fatal %d", 6)

	require.Len(t, formatted.calls, 6)
	assert.Equal(t, logCall{level: "debugf", msg: "trace 1"}, formatted.calls[0])
	assert.Equal(t, logCall{level: "debugf", msg: "debug 2"}, formatted.calls[1])
	assert.Equal(t, logCall{level: "infof", msg: "info 3"}, formatted.calls[2])
	assert.Equal(t, logCall{level: "warnf", msg: "warn 4"}, formatted.calls[3])
	assert.Equal(t, logCall{level: "errorf", msg: "error 5"}, formatted.calls[4])
	assert.Equal(t, logCall{level: "errorf", msg: "fatal 6"}, formatted.calls[5])
}

func TestToFormatted(t *testing.T) {
	logger := &recordingLogger{}
	formatted := ToFormatted(logger)

	formatted.Debugf("debug %d", 1)
	formatted.Infof("info %d", 2)
	formatted.Warnf("warn %d", 3)
	formatted.Errorf("error %d", 4)

	require.Len(t, logger.calls, 4)
	assert.Equal(t, logCall{level: "debug", msg: "debug 1"}, logger.calls[0])
	assert.Equal(t, logCall{level: "info", msg: "info 2"}, logger.calls[1])
	assert.Equal(t, logCall{level: "warn", msg: "warn 3"}, logger.calls[2])
	assert.Equal(t, logCall{level: "error", msg: "error 4"}, logger.calls[3])
}

func TestToFormattedNilSafe(t *testing.T) {
	formatted := ToFormatted(nil)
	require.NotNil(t, formatted)
	assert.NotPanics(t, func() {
		formatted.Debugf("safe")
		formatted.Infof("safe")
		formatted.Warnf("safe")
		formatted.Errorf("safe")
	})
}

func TestPrintfFuncLevelMapping(t *testing.T) {
	testCases := []struct {
		name          string
		level         string
		expectedLevel string
	}{
		{name: "trace", level: "trace", expectedLevel: "trace"},
		{name: "debug", level: "DEBUG", expectedLevel: "debug"},
		{name: "info by default", level: "unknown", expectedLevel: "info"},
		{name: "warn", level: "warn", expectedLevel: "warn"},
		{name: "warning alias", level: "warning", expectedLevel: "warn"},
		{name: "error", level: "error", expectedLevel: "error"},
		{name: "err alias", level: "err", expectedLevel: "error"},
		{name: "fatal", level: "fatal", expectedLevel: "fatal"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := &recordingLogger{}
			printf := PrintfFunc(logger, tc.level)
			printf("value %d", 7)

			require.Len(t, logger.calls, 1)
			assert.Equal(t, tc.expectedLevel, logger.calls[0].level)
			assert.Equal(t, "value 7", logger.calls[0].msg)
		})
	}
}

func TestPrintfFuncNilSafe(t *testing.T) {
	printf := PrintfFunc(nil, "error")
	assert.NotPanics(t, func() {
		printf("safe %s", "message")
	})
}
