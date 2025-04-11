package glog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/fatih/color"
)

type ColorConsoleOptions struct {
	SLogOpts slog.HandlerOptions
}

type ColorConsoleHandler struct {
	slog.Handler
	mu sync.Mutex
}

func NewColorConsoleHandler(out io.Writer, opts *slog.HandlerOptions) *ColorConsoleHandler {
	return &ColorConsoleHandler{
		Handler: slog.NewTextHandler(out, opts),
	}
}

func (c *ColorConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("==== pretty handler handling stuff...")

	level := r.Level.String()

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	case LevelSuccess:
		level = color.GreenString(CustomLevels[LevelSuccess])
	case LevelTrace:
		level = color.WhiteString(CustomLevels[LevelTrace])
	}

	fields := make(map[string]any, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	ts := r.Time.Format("[2006-01-02 15:05:05]")
	msg := color.CyanString(r.Message)

	formattedFields := formatFields(fields, " - ")

	fmt.Printf("%s %s %s %s\n", ts, level, msg, color.WhiteString(formattedFields))

	os.Stdout.Sync()

	return nil
}

func formatFields(fields map[string]any, spacer string) string {
	var result string
	for key, val := range fields {
		result += fmt.Sprintf("%s=%v%s", key, val, spacer)
	}

	if len(result) > 0 {
		result = result[:len(result)-len(spacer)]
	}

	return result
}
