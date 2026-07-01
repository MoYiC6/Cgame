package logger

import (
	"context"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"backend/internal/platform/config"
	"backend/internal/platform/observability"
)

type Field = any

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

type structuredLogger struct {
	inner      *slog.Logger
	sampleRate float64
	seed       int64
	counter    int64
}

func String(key, value string) Field {
	return slog.String(key, value)
}

func Any(key string, value any) Field {
	return slog.Any(key, value)
}

func New(cfg config.LogConfig) Logger {
	return newLogger(cfg.Level, cfg.Format, cfg.SampleRate, os.Stdout)
}

func NewText(level string, writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return newLogger(level, "text", 1.0, writer)
}

func NewWithSample(level, format string, sampleRate float64, writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return newLogger(level, format, sampleRate, writer)
}

func newLogger(level, format string, sampleRate float64, writer io.Writer) Logger {
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

	if sampleRate < 0 {
		sampleRate = 0
	}
	if sampleRate > 1 {
		sampleRate = 1
	}

	opts := &slog.HandlerOptions{Level: slogLevel}
	if slogLevel == slog.LevelDebug {
		opts.AddSource = true
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return &structuredLogger{
		inner:      slog.New(handler),
		sampleRate: sampleRate,
		seed:       time.Now().UnixNano(),
	}
}

func (l *structuredLogger) Debug(msg string, fields ...Field) {
	if l.shouldLog(msg) {
		l.inner.Debug(msg, fields...)
	}
}

func (l *structuredLogger) Info(msg string, fields ...Field) {
	if l.shouldLog(msg) {
		l.inner.Info(msg, fields...)
	}
}

func (l *structuredLogger) Warn(msg string, fields ...Field) {
	l.inner.Warn(msg, fields...)
}

func (l *structuredLogger) Error(msg string, fields ...Field) {
	l.inner.Error(msg, fields...)
}

func (l *structuredLogger) shouldLog(msg string) bool {
	if l.sampleRate >= 1.0 {
		return true
	}
	if l.sampleRate <= 0 {
		return false
	}
	counter := atomic.AddInt64(&l.counter, 1)
	r := rand.New(rand.NewSource(l.seed + counter))
	return r.Float64() < l.sampleRate
}

func (l *structuredLogger) with(fields ...any) Logger {
	return &structuredLogger{inner: l.inner.With(fields...), sampleRate: l.sampleRate, seed: l.seed}
}

func WithContext(ctx context.Context, base Logger) Logger {
	if base == nil {
		base = New(config.LogConfig{Level: "info"})
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