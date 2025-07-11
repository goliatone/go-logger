package glog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var ColorConsoleTSFormat = "2006-01-02 15:04:05.000"

var (
	maxDisplayNameLenMu sync.Mutex
	maxDisplayNameLen   = 6
	maxAllowedNameLen   = 12
)

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
		loggerInfo = h.formatLoggerName(loggerName)
		delete(attrMap, "logger") // remove key from attributes to avoid duplication
	}

	var sourceInfo string
	if h.includeSourceInfo(r.Level) {
		if r.PC != 0 {
			fs := runtime.CallersFrames([]uintptr{r.PC})
			f, _ := fs.Next()
			if f.File != "" {
				source := fmt.Sprintf("%s:%d", f.File, f.Line)
				sourceInfo = color.New(color.FgHiBlack).Sprintf(" (%s)", source)
			}
		}
	}
	// delete(attrMap, "source") //TODO: take source optionally else do this

	var stackInfo string
	if err, ok := attrMap["stack"]; ok {
		stackInfo = color.New(color.FgHiBlack).Sprintf("%s", err)
		delete(attrMap, "stack")

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

	if stackInfo != "" {
		fmt.Fprintf(h.out, "%s", stackInfo)
	}

	return nil
}

func (h *ColorConsoleHandler) updateMaxNameLen(name string) {
	effectiveLen := len(name)
	if effectiveLen > maxAllowedNameLen {
		effectiveLen = maxAllowedNameLen
	}
	maxDisplayNameLenMu.Lock()
	defer maxDisplayNameLenMu.Unlock()

	if effectiveLen > maxDisplayNameLen {
		maxDisplayNameLen = effectiveLen
	}
}

func (h *ColorConsoleHandler) formatLoggerName(name string) string {
	h.updateMaxNameLen(name)

	dislayName := name
	if len(name) > maxAllowedNameLen {
		dislayName = "..." + name[len(name)-(maxAllowedNameLen-3):]
	}

	withBrackets := "[" + dislayName + "]"

	maxDisplayNameLenMu.Lock()
	currentMaxLen := maxDisplayNameLen
	maxDisplayNameLenMu.Unlock()

	// create dyanmic template for display len
	formatStr := fmt.Sprintf("%%%ds", currentMaxLen+2)

	return color.New(color.FgGreen, color.Bold).Sprintf(formatStr, withBrackets)
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

func (h *ColorConsoleHandler) includeSourceInfo(level slog.Level) bool {
	switch level {
	case slog.LevelError:
		return true
	case slog.LevelWarn:
		return true
	default:
		return h.opts.AddSource
	}
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
	levelName = fmt.Sprintf("%-6s", levelName)

	// Apply color based on level
	switch level {
	case LevelTrace:
		return color.New(color.FgHiBlack).Sprint(levelName)
	case slog.LevelDebug:
		return color.New(color.FgMagenta).Sprint(levelName)
	case slog.LevelInfo:
		return color.New(color.FgBlue).Sprint(levelName)
	case slog.LevelWarn:
		return color.New(color.FgYellow).Sprint(levelName)
	case slog.LevelError:
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

	var keys []string
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := attrs[k]
		var keyColored string
		if k == "error" {
			keyColored = color.New(color.FgHiRed).Sprint("message")
		} else {
			keyColored = color.New(color.FgHiYellow).Sprint(k)
		}
		parts = append(parts, fmt.Sprintf(" %s=%v", keyColored, v))
	}

	return strings.Join(parts, "")
}
