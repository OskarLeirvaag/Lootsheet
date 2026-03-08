package app

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	testCases := []struct {
		input string
		want  slog.Level
	}{
		{input: "", want: slog.LevelInfo},
		{input: "INFO", want: slog.LevelInfo},
		{input: "DBG", want: slog.LevelDebug},
		{input: "debug", want: slog.LevelDebug},
		{input: "WARN", want: slog.LevelWarn},
		{input: "ERR", want: slog.LevelError},
		{input: "unknown", want: slog.LevelInfo},
	}

	for _, testCase := range testCases {
		if got := parseLogLevel(testCase.input); got != testCase.want {
			t.Fatalf("parseLogLevel(%q) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
}

func TestFormatLevel(t *testing.T) {
	testCases := []struct {
		level slog.Level
		want  string
	}{
		{level: slog.LevelDebug, want: "DBG"},
		{level: slog.LevelInfo, want: "INFO"},
		{level: slog.LevelWarn, want: "WARN"},
		{level: slog.LevelError, want: "ERR"},
	}

	for _, testCase := range testCases {
		got := formatLevel(slog.AnyValue(testCase.level))
		if got != testCase.want {
			t.Fatalf("formatLevel(%v) = %q, want %q", testCase.level, got, testCase.want)
		}
	}
}
