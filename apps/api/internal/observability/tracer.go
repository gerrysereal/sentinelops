package observability

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// SDK owns the process-wide OpenTelemetry providers. It is safe to construct
// once during application bootstrap and shut down during graceful shutdown.
type SDK struct {
	serviceName    string
	tracerProvider trace.TracerProvider
	meterProvider  MeterProvider
	propagator     propagation.TextMapPropagator
	shutdown       []func(context.Context) error
	shutdownOnce   sync.Once
	shutdownErr    error
}

// New initializes trace and metric providers and installs them as the global
// OpenTelemetry providers. When telemetry is disabled, standards-compliant
// no-op providers are installed without opening network connections.
func New(ctx context.Context, cfg Config) (*SDK, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate observability configuration: %w", err)
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	if !cfg.Enabled {
		tracerProvider := tracenoop.NewTracerProvider()
		meterProvider := newNoopMeterProvider()
		otel.SetTracerProvider(tracerProvider)
		otel.SetMeterProvider(meterProvider)
		otel.SetTextMapPropagator(propagator)
		return &SDK{
			serviceName:    cfg.ServiceName,
			tracerProvider: tracerProvider,
			meterProvider:  meterProvider,
			propagator:     propagator,
		}, nil
	}

	res, err := newResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tracerProvider, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return nil, err
	}

	meterProvider, err := newMeterProvider(ctx, cfg, res)
	if err != nil {
		_ = tracerProvider.Shutdown(ctx)
		return nil, err
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagator)

	return &SDK{
		serviceName:    cfg.ServiceName,
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		propagator:     propagator,
		shutdown: []func(context.Context) error{
			meterProvider.Shutdown,
			tracerProvider.Shutdown,
		},
	}, nil
}

// TracerProvider exposes the configured provider to instrumentation adapters
// without coupling callers to the SDK implementation type.
func (s *SDK) TracerProvider() trace.TracerProvider {
	return s.tracerProvider
}

// Propagator exposes the W3C Trace Context and Baggage propagator used by the
// service so inbound and outbound instrumentation share the same context.
func (s *SDK) Propagator() propagation.TextMapPropagator {
	return s.propagator
}

// ForceFlush immediately exports buffered spans and metrics. It is intended
// for controlled shutdowns and operational diagnostics, not per-request use.
func (s *SDK) ForceFlush(ctx context.Context) error {
	var flushErrors []error
	if provider, ok := s.tracerProvider.(*sdktrace.TracerProvider); ok {
		if err := provider.ForceFlush(ctx); err != nil {
			flushErrors = append(flushErrors, fmt.Errorf("flush traces: %w", err))
		}
	}
	if err := forceFlushMetrics(ctx, s.meterProvider); err != nil {
		flushErrors = append(flushErrors, err)
	}
	return errors.Join(flushErrors...)
}

// Shutdown flushes and closes all exporters exactly once.
func (s *SDK) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		var shutdownErrors []error
		for _, shutdown := range s.shutdown {
			if err := shutdown(ctx); err != nil {
				shutdownErrors = append(shutdownErrors, err)
			}
		}
		s.shutdownErr = errors.Join(shutdownErrors...)
	})
	return s.shutdownErr
}

func newTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	endpoint, insecure, err := normalizeGRPCEndpoint(cfg.OTLPEndpoint, cfg.OTLPInsecure)
	if err != nil {
		return nil, fmt.Errorf("configure OTLP trace exporter: %w", err)
	}

	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithTimeout(cfg.ExportTimeout),
	}
	if insecure {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	sampler, err := samplerFromConfig(cfg.Sampler, cfg.SamplerArgument)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
			sdktrace.WithExportTimeout(cfg.ExportTimeout),
		),
	), nil
}

func samplerFromConfig(name string, argument float64) (sdktrace.Sampler, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "always_on":
		return sdktrace.AlwaysSample(), nil
	case "always_off":
		return sdktrace.NeverSample(), nil
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(argument), nil
	case "parentbased_always_on", "":
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample()), nil
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(argument)), nil
	default:
		return nil, fmt.Errorf("unsupported OTEL_TRACES_SAMPLER value %q", name)
	}
}

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	attributes := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
	}
	if cfg.ServiceVersion != "" {
		attributes = append(attributes, attribute.String("service.version", cfg.ServiceVersion))
	}
	if cfg.DeploymentEnvironment != "" {
		attributes = append(attributes, attribute.String("deployment.environment.name", cfg.DeploymentEnvironment))
	}
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		attributes = append(attributes, attribute.String("service.instance.id", hostname))
	}

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attributes...),
	)
	if err != nil {
		return nil, fmt.Errorf("create OpenTelemetry resource: %w", err)
	}
	return res, nil
}
