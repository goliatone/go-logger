package glog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var ColorConsoleTSFormat = "2006-01-02 15:04:05.000"

type ColorConsoleOption func(*ColorConsoleHandler)

func WithColorConsoleTSFormat(format string) ColorConsoleOption {
	return func(cch *ColorConsoleHandler) {
		cch.tsFormat = format
	}
}

// ColorConsoleHandler is a custom slog.Handler that outputs colored logs to the console
type ColorConsoleHandler struct {
	out      io.Writer
	opts     *slog.HandlerOptions
	mu       *sync.Mutex
	attrs    []slog.Attr
	groups   []string
	tsFormat string
}

// NewColorConsoleHandler creates a new ColorConsoleHandler with the provided options
func NewColorConsoleHandler(out io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &ColorConsoleHandler{
		out:      out,
		opts:     opts,
		mu:       &sync.Mutex{},
		attrs:    []slog.Attr{},
		groups:   []string{},
		tsFormat: ColorConsoleTSFormat,
	}
}

func (h *ColorConsoleHandler) WithTSFormat(format string) *ColorConsoleHandler {
	h.tsFormat = format
	return h
}

// Enabled implements slog.Handler.
func (h *ColorConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle implements slog.Handler.
func (h *ColorConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	coloredLevel := h.colorizeLevel(r.Level)

	ts := r.Time.Format(h.tsFormat)
	coloredTs := color.New(color.FgHiBlack).Sprint(ts)

	msg := r.Message
	coloredMsg := color.WhiteString(msg)

	attrMap := make(map[string]any)

	for _, attr := range h.attrs {
		attrMap[attr.Key] = attr.Value.Any()
	}

	r.Attrs(func(a slog.Attr) bool {
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(h.groups, a)
		}

		if a.Equal(slog.Attr{}) {
			return true
		}

		key := a.Key
		if len(h.groups) > 0 {
			key = strings.Join(append(slices.Clone(h.groups), key), ".")
		}

		attrMap[key] = a.Value.Any()
		return true
	})

	var loggerInfo string
	if loggerName, ok := attrMap["logger"].(string); ok {
		loggerName = "[" + loggerName + "]"
		loggerInfo = color.New(color.FgGreen, color.Bold).Sprintf("%6s", loggerName)
		delete(attrMap, "logger") // remove key from attributes to avoid duplication
	}

	var sourceInfo string
	if source, ok := attrMap["source"]; ok && h.opts.AddSource {
		sourceInfo = color.New(color.FgHiBlack).Sprintf("(%s)", source)
		delete(attrMap, "source")
	}

	delete(attrMap, "ts")
	delete(attrMap, "time")
	delete(attrMap, "level")

	var formattedAttrs string
	if len(attrMap) > 0 {
		formattedAttrs = h.formatAttrs(attrMap)
	}

	// TODO: can we use a template here?
	fmt.Fprintf(h.out, "%s %s %s%s %s %s\n",
		loggerInfo,
		coloredTs,
		coloredLevel,
		coloredMsg,
		formattedAttrs,
		sourceInfo,
	)

	return nil
}

// WithAttrs implements slog.Handler.
func (h *ColorConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = append(slices.Clone(h.attrs), attrs...)
	return &h2
}

// WithGroup implements slog.Handler.
func (h *ColorConsoleHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	h2.groups = append(slices.Clone(h.groups), name)
	return &h2
}

// colorizeLevel returns the level string with appropriate color
func (h *ColorConsoleHandler) colorizeLevel(level slog.Level) string {
	levelName := level.String()

	// Check for custom level names
	if customName, exists := CustomLevels[level]; exists {
		levelName = customName
	}

	// Make it uppercase and pad it for alignment
	levelName = strings.ToUpper(levelName)
	levelName = fmt.Sprintf("%-5s", levelName)

	// Apply color based on level
	switch {
	case level == LevelTrace:
		return color.New(color.FgHiBlack).Sprint(levelName)
	case level == slog.LevelDebug:
		return color.New(color.FgMagenta).Sprint(levelName)
	case level == slog.LevelInfo:
		return color.New(color.FgBlue).Sprint(levelName)
	case level == slog.LevelWarn:
		return color.New(color.FgYellow).Sprint(levelName)
	case level == slog.LevelError:
		return color.New(color.FgRed, color.Bold).Sprint(levelName)
	default:
		return levelName
	}
}

// formatAttrs formats a map of attributes into a string
func (h *ColorConsoleHandler) formatAttrs(attrs map[string]any) string {
	if len(attrs) == 0 {
		return ""
	}

	var parts []string
	for k, v := range attrs {
		// Format the key-value pair
		key := color.New(color.FgHiYellow).Sprint(k)
		val := fmt.Sprintf("%v", v)
		parts = append(parts, fmt.Sprintf(" %s=%s", key, val))
	}

	return strings.Join(parts, "")
}
