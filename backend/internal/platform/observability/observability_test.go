package observability

import (
	"context"
	"errors"
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

func TestNoopTracerStartSpanReturnsSameContext(t *testing.T) {
	tracer := NewNoopTracer()
	ctx := WithRequestID(context.Background(), "req-1")

	gotCtx, span := tracer.StartSpan(ctx, "demo-span")

	if gotCtx != ctx {
		t.Fatalf("expected StartSpan to return same context")
	}

	if span == nil {
		t.Fatal("expected noop span, got nil")
	}
}

func TestNoopSpanEndDoesNotPanic(t *testing.T) {
	_, noopSpan := NewNoopTracer().StartSpan(context.Background(), "demo")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected End not to panic, got %v", r)
		}
	}()

	noopSpan.End(nil)
	noopSpan.End(errors.New("boom"))
}

func TestNoopPropagatorInjectNoTraceID(t *testing.T) {
	propagator := NewNoopPropagator()
	carrier := MapCarrier{}

	propagator.Inject(context.Background(), carrier)

	if len(carrier) != 0 {
		t.Fatalf("expected carrier to remain empty, got %v", carrier)
	}
	if _, ok := carrier["X-Trace-ID"]; ok {
		t.Fatal("expected no X-Trace-ID to be written")
	}
}

func TestNoopPropagatorExtractWithoutHeaderReturnsOriginalContext(t *testing.T) {
	propagator := NewNoopPropagator()
	ctx := WithRequestID(context.Background(), "req-2")
	carrier := MapCarrier{}

	gotCtx := propagator.Extract(ctx, carrier)

	if gotCtx != ctx {
		t.Fatal("expected Extract to return original context when header is absent")
	}
	if traceID, ok := TraceIDFromContext(gotCtx); ok || traceID != "" {
		t.Fatalf("expected no trace id, got %q", traceID)
	}
}

func TestNoopPropagatorExtractWithHeaderAddsTraceID(t *testing.T) {
	propagator := NewNoopPropagator()
	ctx := context.Background()
	carrier := MapCarrier{"X-Trace-ID": "trace-abc"}

	gotCtx := propagator.Extract(ctx, carrier)

	traceID, ok := TraceIDFromContext(gotCtx)
	if !ok || traceID != "trace-abc" {
		t.Fatalf("expected extracted trace id trace-abc, got %q", traceID)
	}
}

func TestMapCarrierKeysContainsAllInsertedKeys(t *testing.T) {
	carrier := MapCarrier{}
	carrier.Set("alpha", "1")
	carrier.Set("beta", "2")

	if got := carrier.Get("alpha"); got != "1" {
		t.Fatalf("expected alpha to be 1, got %q", got)
	}
	if got := carrier.Get("beta"); got != "2" {
		t.Fatalf("expected beta to be 2, got %q", got)
	}

	keys := carrier.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}

	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		seen[key] = struct{}{}
	}

	if _, ok := seen["alpha"]; !ok {
		t.Fatal("expected alpha in keys")
	}
	if _, ok := seen["beta"]; !ok {
		t.Fatal("expected beta in keys")
	}
	if len(seen) != 2 {
		t.Fatalf("expected unique keys, got %v", keys)
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
