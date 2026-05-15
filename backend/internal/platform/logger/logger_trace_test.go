package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"backend/internal/platform/observability"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestLoggerWithContextIncludesTraceID(t *testing.T) {
	var buffer bytes.Buffer
	base := New("info", &buffer)

	traceID, err := oteltrace.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("TraceIDFromHex returned error: %v", err)
	}
	spanID, err := oteltrace.SpanIDFromHex("1234567890abcdef")
	if err != nil {
		t.Fatalf("SpanIDFromHex returned error: %v", err)
	}

	ctx := oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
		Remote:  false,
	}))

	WithContext(ctx, base).Info("trace linked")

	output := buffer.String()
	if !strings.Contains(output, "trace_id=1234567890abcdef1234567890abcdef") {
		t.Fatalf("expected output to contain otel trace id, got %s", output)
	}
}

func TestLoggerWithContextFallsBackToCustomTraceID(t *testing.T) {
	var buffer bytes.Buffer
	base := New("info", &buffer)

	ctx := observability.WithTraceID(context.Background(), "custom-id")

	WithContext(ctx, base).Info("trace linked")

	output := buffer.String()
	if !strings.Contains(output, "trace_id=custom-id") {
		t.Fatalf("expected output to contain custom trace id, got %s", output)
	}
}
