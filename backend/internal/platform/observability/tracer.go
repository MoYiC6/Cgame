package observability

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

type Span interface {
	End(err error)
}

type noopTracer struct{}

type noopSpan struct{}

type otelTracer struct {
	inner oteltrace.Tracer
}

type otelSpan struct {
	inner oteltrace.Span
}

func NewNoopTracer() Tracer {
	return noopTracer{}
}

func (noopTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, noopSpan{}
}

func (noopSpan) End(err error) {}

func newOTELTracer(inner oteltrace.Tracer) Tracer {
	if inner == nil {
		return NewNoopTracer()
	}
	return otelTracer{inner: inner}
}

func (t otelTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	ctx, span := t.inner.Start(ctx, name)
	if spanContext := span.SpanContext(); spanContext.IsValid() {
		ctx = WithTraceID(ctx, spanContext.TraceID().String())
	}
	return ctx, otelSpan{inner: span}
}

func (s otelSpan) End(err error) {
	if s.inner == nil {
		return
	}
	if err != nil {
		s.inner.RecordError(err)
	}
	s.inner.End()
}
