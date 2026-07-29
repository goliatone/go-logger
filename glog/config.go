package glog

import "strings"

// NormalizeLevel returns the canonical log-level constant for value.
// Matching is case-insensitive and ignores surrounding whitespace.
// Unsupported and empty values fall back to Info.
func NormalizeLevel(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case Trace:
		return Trace
	case Debug:
		return Debug
	case Info:
		return Info
	case Warn, "WARNING":
		return Warn
	case Error:
		return Error
	case Fatal:
		return Fatal
	default:
		return Info
	}
}

// NormalizeLoggerType returns the canonical output type for value.
// Matching is case-insensitive and ignores surrounding whitespace.
// Text is accepted as an alias for console. Unsupported and empty values
// fall back to JSON.
func NormalizeLoggerType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case LoggerTypeConsole, "text":
		return LoggerTypeConsole
	case LoggerTypePretty:
		return LoggerTypePretty
	case LoggerTypeJSON:
		return LoggerTypeJSON
	default:
		return LoggerTypeJSON
	}
}
