package glog

import "testing"

func TestNormalizeLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trace", input: "trace", want: Trace},
		{name: "debug mixed case", input: "DeBuG", want: Debug},
		{name: "info padded", input: "  info\t", want: Info},
		{name: "warn", input: "WARN", want: Warn},
		{name: "warning alias", input: " warning ", want: Warn},
		{name: "error", input: "error", want: Error},
		{name: "fatal", input: "fatal", want: Fatal},
		{name: "empty fallback", input: "", want: Info},
		{name: "unsupported fallback", input: "verbose", want: Info},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeLevel(tt.input); got != tt.want {
				t.Fatalf("NormalizeLevel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeLoggerType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "json", input: "json", want: LoggerTypeJSON},
		{name: "console mixed case", input: "CoNsOlE", want: LoggerTypeConsole},
		{name: "text alias", input: " text ", want: LoggerTypeConsole},
		{name: "pretty padded", input: "\tpretty ", want: LoggerTypePretty},
		{name: "empty fallback", input: "", want: LoggerTypeJSON},
		{name: "unsupported fallback", input: "xml", want: LoggerTypeJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeLoggerType(tt.input); got != tt.want {
				t.Fatalf("NormalizeLoggerType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
