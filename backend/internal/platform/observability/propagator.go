package observability

import "context"

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

func NewNoopPropagator() Propagator {
	return noopPropagator{}
}

func (m MapCarrier) Get(key string) string {
	return m[key]
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
	}
}

func (noopPropagator) Extract(ctx context.Context, carrier Carrier) context.Context {
	if traceID := carrier.Get("X-Trace-ID"); traceID != "" {
		return WithTraceID(ctx, traceID)
	}
	return ctx
}
