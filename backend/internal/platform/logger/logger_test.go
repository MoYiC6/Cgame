package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"backend/internal/platform/observability"
)

func TestWithContextAddsRequestAndTraceIDs(t *testing.T) {
	var buffer bytes.Buffer
	base := New("debug", &buffer)
	ctx := context.Background()
	ctx = observability.WithRequestID(ctx, "req-log")
	ctx = observability.WithTraceID(ctx, "trace-log")

	WithContext(ctx, base).Info("boot ok", "component", "api")

	output := buffer.String()
	if !strings.Contains(output, "req-log") {
		t.Fatalf("expected output to contain request id, got %s", output)
	}
	if !strings.Contains(output, "trace-log") {
		t.Fatalf("expected output to contain trace id, got %s", output)
	}
}
