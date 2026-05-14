package observability

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/platform/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Provider interface {
	Tracer() Tracer
	Propagator() Propagator
	Shutdown(ctx context.Context) error
	IsNoop() bool
}

type provider struct {
	tracer     Tracer
	propagator Propagator
	shutdown   func(context.Context) error
	noop       bool
}

func InitProvider(ctx context.Context, cfg config.ObservabilityConfig) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.TraceExporterType)) {
	case "", "none":
		return newNoopProvider(), nil
	case "otlp":
		return initOTLPProvider(ctx, cfg)
	default:
		return newNoopProvider(), fmt.Errorf("unsupported trace exporter type %q", cfg.TraceExporterType)
	}
}

func newNoopProvider() Provider {
	return &provider{
		tracer:     NewNoopTracer(),
		propagator: NewNoopPropagator(),
		shutdown:   func(context.Context) error { return nil },
		noop:       true,
	}
}

func initOTLPProvider(ctx context.Context, cfg config.ObservabilityConfig) (Provider, error) {
	endpoint := strings.TrimSpace(cfg.TraceExporterEndpoint)
	if endpoint == "" {
		return newNoopProvider(), fmt.Errorf("otlp exporter endpoint is required")
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return newNoopProvider(), fmt.Errorf("initialize otlp trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(strings.TrimSpace(cfg.ServiceName)),
			semconv.ServiceVersion(strings.TrimSpace(cfg.ServiceVersion)),
			attribute.String("deployment.environment", strings.TrimSpace(cfg.Environment)),
		),
	)
	if err != nil {
		shutdownCtx := context.Background()
		_ = exporter.Shutdown(shutdownCtx)
		return newNoopProvider(), fmt.Errorf("build otel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	textMapPropagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(textMapPropagator)

	return &provider{
		tracer:     newOTELTracer(tp.Tracer(strings.TrimSpace(cfg.ServiceName))),
		propagator: newOTELPropagator(textMapPropagator),
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
		noop: false,
	}, nil
}

func (p *provider) Tracer() Tracer {
	if p == nil || p.tracer == nil {
		return NewNoopTracer()
	}
	return p.tracer
}

func (p *provider) Propagator() Propagator {
	if p == nil || p.propagator == nil {
		return NewNoopPropagator()
	}
	return p.propagator
}

func (p *provider) Shutdown(ctx context.Context) error {
	if p == nil || p.shutdown == nil {
		return nil
	}
	return p.shutdown(ctx)
}

func (p *provider) IsNoop() bool {
	if p == nil {
		return true
	}
	return p.noop
}
