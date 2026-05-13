package observability

import (
	"context"
	"testing"
)

func TestRequestAndTraceContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-42")
	ctx = WithTraceID(ctx, "trace-42")

	requestID, ok := RequestIDFromContext(ctx)
	if !ok || requestID != "req-42" {
		t.Fatalf("expected request id req-42, got %q", requestID)
	}

	traceID, ok := TraceIDFromContext(ctx)
	if !ok || traceID != "trace-42" {
		t.Fatalf("expected trace id trace-42, got %q", traceID)
	}
}

func TestNoopPropagatorInjectExtract(t *testing.T) {
	propagator := NewNoopPropagator()
	carrier := MapCarrier{}
	ctx := WithTraceID(context.Background(), "trace-prop")

	propagator.Inject(ctx, carrier)
	newCtx := propagator.Extract(context.Background(), carrier)

	traceID, ok := TraceIDFromContext(newCtx)
	if !ok || traceID != "trace-prop" {
		t.Fatalf("expected extracted trace id trace-prop, got %q", traceID)
	}
}
