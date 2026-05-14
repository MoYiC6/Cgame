package observability

import (
	"context"
	"strings"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(requestIDKey).(string)
	return value, ok && value != ""
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	if spanContext := oteltrace.SpanContextFromContext(ctx); spanContext.IsValid() {
		traceID := spanContext.TraceID().String()
		if strings.TrimSpace(traceID) != "" {
			return traceID, true
		}
	}
	value, ok := ctx.Value(traceIDKey).(string)
	return value, ok && value != ""
}
