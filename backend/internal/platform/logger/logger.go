package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"backend/internal/platform/observability"
)

type Field = any

type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

func String(key, value string) Field {
	return slog.String(key, value)
}

func Any(key string, value any) Field {
	return slog.Any(key, value)
}

func New(level string, writer io.Writer) *slog.Logger {
	if writer == nil {
		writer = os.Stdout
	}

	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler)
}

func WithContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if base == nil {
		base = New("info", nil)
	}

	fields := make([]any, 0, 4)
	if requestID, ok := observability.RequestIDFromContext(ctx); ok {
		fields = append(fields, "request_id", requestID)
	}
	if traceID, ok := observability.TraceIDFromContext(ctx); ok {
		fields = append(fields, "trace_id", traceID)
	}
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}
