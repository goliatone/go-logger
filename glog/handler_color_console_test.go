package glog

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestColorConsoleHandler_Handle(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}
	handler := NewColorConsoleHandler(&buf, opts)
	logger := slog.New(handler)

	logger.Info("user logged in", slog.String("user", "test"), slog.Int("id", 123))

	output := buf.String()
	fmt.Println(output)

	assert.Regexp(t, regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}`), output)

	assert.Contains(t, output, "INFO  ")

	assert.Contains(t, output, "user logged in")

	assert.Regexp(t, regexp.MustCompile(`id=123`), output)
	assert.Regexp(t, regexp.MustCompile(`user=test`), output)

	assert.Regexp(t, regexp.MustCompile(`\(.*/glog/handler_color_console_test.go:\d+\)`), output)
}

func TestColorConsoleHandler_WithGroupAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := NewColorConsoleHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

	handler = handler.WithAttrs([]slog.Attr{slog.String("request_id", "abc")})
	handler = handler.WithGroup("user")

	record := slog.NewRecord(time.Now(), slog.LevelWarn, "update profile", 0)
	record.AddAttrs(slog.String("name", "john"))

	err := handler.Handle(context.Background(), record)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "WARN  ")
	assert.Contains(t, output, "update profile")
	assert.Regexp(t, regexp.MustCompile(`request_id=abc`), output)
	assert.Regexp(t, regexp.MustCompile(`user\.name=john`), output)
}

func TestColorConsoleHandler_LoggerNameFormatting(t *testing.T) {
	t.Cleanup(func() {
		maxDisplayNameLenMu.Lock()
		maxDisplayNameLen = 6
		maxDisplayNameLenMu.Unlock()
	})

	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := NewColorConsoleHandler(&buf, opts)

	record1 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg1", 0)
	record1.AddAttrs(slog.String("logger", "api"))
	handler.Handle(context.Background(), record1)
	output1 := buf.String()
	assert.Contains(t, output1, "[api] ")

	buf.Reset()
	record2 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg2", 0)
	record2.AddAttrs(slog.String("logger", "database"))
	handler.Handle(context.Background(), record2)
	output2 := buf.String()
	assert.Contains(t, output2, "[database] ")

	buf.Reset()
	record3 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg3", 0)
	record3.AddAttrs(slog.String("logger", "very-long-service-name"))
	handler.Handle(context.Background(), record3)
	output3 := buf.String()

	assert.Regexp(t, regexp.MustCompile(`\[\.\.\.vice-name\]\s+`), output3)
}

func TestColorizeLevel(t *testing.T) {
	handler := &ColorConsoleHandler{}

	testCases := []struct {
		level    slog.Level
		expected string
	}{
		{LevelTrace, "TRACE "},
		{slog.LevelDebug, "DEBUG "},
		{slog.LevelInfo, "INFO  "},
		{slog.LevelWarn, "WARN  "},
		{slog.LevelError, "ERROR "},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			colored := handler.colorizeLevel(tc.level)
			uncolored := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(colored, "")
			assert.Equal(t, tc.expected, uncolored)
		})
	}
}

func TestFormatAttrs(t *testing.T) {
	handler := &ColorConsoleHandler{}

	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", handler.formatAttrs(map[string]any{}))
	})

	t.Run("sorted with values", func(t *testing.T) {
		attrs := map[string]any{
			"c_key": 123,
			"a_key": "hello",
			"b_key": true,
		}
		formatted := handler.formatAttrs(attrs)
		uncolored := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(formatted, "")
		assert.Equal(t, " a_key=hello b_key=true c_key=123", uncolored)
	})

	t.Run("error key renamed", func(t *testing.T) {
		attrs := map[string]any{
			"error":   "not found",
			"request": "123",
		}
		formatted := handler.formatAttrs(attrs)
		uncolored := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(formatted, "")
		assert.True(t, strings.HasPrefix(uncolored, " message=not found"))
		assert.True(t, strings.HasSuffix(uncolored, " request=123"))
	})
}
