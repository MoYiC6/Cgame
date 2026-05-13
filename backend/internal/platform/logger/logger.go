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

type structuredLogger struct {
	inner *slog.Logger
}

func String(key, value string) Field {
	return slog.String(key, value)
}

func Any(key string, value any) Field {
	return slog.Any(key, value)
}

func New(level string, writer io.Writer) Logger {
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

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slogLevel})
	return &structuredLogger{inner: slog.New(handler)}
}

func (l *structuredLogger) Info(msg string, fields ...Field) {
	l.inner.Info(msg, fields...)
}

func (l *structuredLogger) Error(msg string, fields ...Field) {
	l.inner.Error(msg, fields...)
}

func (l *structuredLogger) with(fields ...any) Logger {
	return &structuredLogger{inner: l.inner.With(fields...)}
}

func WithContext(ctx context.Context, base Logger) Logger {
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
	contextual, ok := base.(interface{ with(fields ...any) Logger })
	if !ok {
		return base
	}
	return contextual.with(fields...)
}
