package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

const (
	defaultLogLevel   = "INFO"
	defaultLoggerName = "github.com/OskarLeirvaag/Lootsheet"
	logLevelEnvVar    = "LOOTSHEET_LOG_LEVEL"
)

type appLogger struct {
	logger   *slog.Logger
	shutdown func(context.Context) error
}

func newAppLogger(output io.Writer) (*appLogger, error) {
	if output == nil {
		return nil, fmt.Errorf("log output is required")
	}

	levelText := strings.TrimSpace(os.Getenv(logLevelEnvVar))
	if levelText == "" {
		levelText = defaultLogLevel
	}

	level := parseLogLevel(levelText)
	provider := sdklog.NewLoggerProvider()
	logglobal.SetLoggerProvider(provider)

	consoleHandler := slog.NewTextHandler(output, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceLogAttr,
	})
	otelHandler := otelslog.NewHandler(defaultLoggerName, otelslog.WithLoggerProvider(provider))

	logger := slog.New(newMultiHandler(consoleHandler, otelHandler)).With(
		slog.String("app", "lootsheet"),
	)

	return &appLogger{
		logger: logger,
		shutdown: func(ctx context.Context) error {
			return provider.Shutdown(ctx)
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

type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) slog.Handler {
	filtered := make([]slog.Handler, 0, len(handlers))
	for _, handler := range handlers {
		if handler != nil {
			filtered = append(filtered, handler)
		}
	}

	return &multiHandler{handlers: filtered}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error { //nolint:gocritic // slog.Handler interface requires value receiver for Record
	var firstErr error

	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}

		if err := handler.Handle(ctx, record.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithAttrs(attrs))
	}

	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler.WithGroup(name))
	}

	return &multiHandler{handlers: handlers}
}
