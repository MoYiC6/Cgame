package observability

import "context"

type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

type Span interface {
	End(err error)
}

type noopTracer struct{}

type noopSpan struct{}

func NewNoopTracer() Tracer {
	return noopTracer{}
}

func (noopTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, noopSpan{}
}

func (noopSpan) End(err error) {}
