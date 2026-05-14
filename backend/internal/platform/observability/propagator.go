package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"go.opentelemetry.io/otel/propagation"
)

type Propagator interface {
	Inject(ctx context.Context, carrier Carrier)
	Extract(ctx context.Context, carrier Carrier) context.Context
}

type Carrier interface {
	Get(key string) string
	Set(key string, value string)
	Keys() []string
}

type MapCarrier map[string]string

type noopPropagator struct{}

type otelPropagator struct {
	inner propagation.TextMapPropagator
}

func NewNoopPropagator() Propagator {
	return noopPropagator{}
}

func (m MapCarrier) Get(key string) string {
	if value := m[key]; value != "" {
		return value
	}
	for carrierKey, value := range m {
		if strings.EqualFold(carrierKey, key) {
			return value
		}
	}
	return ""
}

func (m MapCarrier) Set(key string, value string) {
	m[key] = value
}

func (m MapCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func (noopPropagator) Inject(ctx context.Context, carrier Carrier) {
	if traceID, ok := TraceIDFromContext(ctx); ok {
		carrier.Set("X-Trace-ID", traceID)
		if carrier.Get("traceparent") == "" && isW3CTraceID(traceID) {
			carrier.Set("traceparent", "00-"+traceID+"-"+newSpanID()+"-01")
		}
	}
}

func (noopPropagator) Extract(ctx context.Context, carrier Carrier) context.Context {
	if traceID := traceIDFromTraceparent(carrier.Get("traceparent")); traceID != "" {
		return WithTraceID(ctx, traceID)
	}
	if traceID := carrier.Get("X-Trace-ID"); traceID != "" {
		return WithTraceID(ctx, traceID)
	}
	return ctx
}

func newOTELPropagator(inner propagation.TextMapPropagator) Propagator {
	if inner == nil {
		return NewNoopPropagator()
	}
	return otelPropagator{inner: inner}
}

func (p otelPropagator) Inject(ctx context.Context, carrier Carrier) {
	p.inner.Inject(ctx, propagationCarrier{carrier: carrier})
	noopPropagator{}.Inject(ctx, carrier)
}

func (p otelPropagator) Extract(ctx context.Context, carrier Carrier) context.Context {
	ctx = p.inner.Extract(ctx, propagationCarrier{carrier: carrier})
	if traceID, ok := TraceIDFromContext(ctx); ok {
		return WithTraceID(ctx, traceID)
	}
	return noopPropagator{}.Extract(ctx, carrier)
}

type propagationCarrier struct {
	carrier Carrier
}

func (c propagationCarrier) Get(key string) string {
	return c.carrier.Get(key)
}

func (c propagationCarrier) Set(key string, value string) {
	c.carrier.Set(key, value)
}

func (c propagationCarrier) Keys() []string {
	return c.carrier.Keys()
}

func traceIDFromTraceparent(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 4 {
		return ""
	}
	traceID := strings.ToLower(parts[1])
	if !isW3CTraceID(traceID) {
		return ""
	}
	return traceID
}

func isW3CTraceID(traceID string) bool {
	if len(traceID) != 32 || traceID == "00000000000000000000000000000000" {
		return false
	}
	_, err := hex.DecodeString(traceID)
	return err == nil
}

func newSpanID() string {
	var id [8]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "0000000000000001"
	}
	return hex.EncodeToString(id[:])
}
