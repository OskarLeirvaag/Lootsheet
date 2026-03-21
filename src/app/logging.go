package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	defaultLogLevel = "INFO"
	logLevelEnvVar  = "LOOTSHEET_LOG_LEVEL"
)

type appLogger struct {
	logger   *slog.Logger
	shutdown func(context.Context) error
}

func newAppLogger(output io.Writer) (*appLogger, error) {
	if output == nil {
		return nil, errors.New("log output is required")
	}

	levelText := strings.TrimSpace(os.Getenv(logLevelEnvVar))
	if levelText == "" {
		levelText = defaultLogLevel
	}

	level := parseLogLevel(levelText)

	handler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceLogAttr,
	})

	logger := slog.New(handler).With(
		slog.String("app", "lootsheet"),
	)

	return &appLogger{
		logger: logger,
		shutdown: func(_ context.Context) error {
			return nil
		},
	}, nil
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToUpper(value) {
	case "DBG", "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERR", "ERROR":
		return slog.LevelError
	case "", "INFO":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

func replaceLogAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key != slog.LevelKey {
		return attr
	}

	return slog.String(attr.Key, formatLevel(attr.Value))
}

func formatLevel(value slog.Value) string {
	if level, ok := value.Any().(slog.Level); ok {
		switch {
		case level <= slog.LevelDebug:
			return "DBG"
		case level < slog.LevelWarn:
			return "INFO"
		case level < slog.LevelError:
			return "WARN"
		default:
			return "ERR"
		}
	}

	switch strings.ToUpper(value.String()) {
	case "DEBUG":
		return "DBG"
	case "WARN", "WARNING":
		return "WARN"
	case "ERROR":
		return "ERR"
	case "INFO":
		return "INFO"
	default:
		return value.String()
	}
}
