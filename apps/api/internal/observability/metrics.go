package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MeterProvider is the narrow contract required by SentinelOps HTTP and data
// store instrumentation. The SDK implementation and the no-op implementation
// both satisfy it.
type MeterProvider interface {
	metric.MeterProvider
}

// MeterProvider exposes the configured provider to instrumentation adapters.
func (s *SDK) MeterProvider() metric.MeterProvider {
	return s.meterProvider
}

func newMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	endpoint, insecure, err := normalizeGRPCEndpoint(cfg.OTLPEndpoint, cfg.OTLPInsecure)
	if err != nil {
		return nil, fmt.Errorf("configure OTLP metric exporter: %w", err)
	}

	exporterOptions := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithTimeout(cfg.ExportTimeout),
	}
	if insecure {
		exporterOptions = append(exporterOptions, otlpmetricgrpc.WithInsecure())
	}
	if len(cfg.OTLPHeaders) > 0 {
		exporterOptions = append(exporterOptions, otlpmetricgrpc.WithHeaders(cfg.OTLPHeaders))
	}

	exporter, err := otlpmetricgrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(cfg.MetricExportInterval),
		sdkmetric.WithTimeout(cfg.ExportTimeout),
	)

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	), nil
}

func newNoopMeterProvider() metric.MeterProvider {
	return metricnoop.NewMeterProvider()
}

func forceFlushMetrics(ctx context.Context, provider MeterProvider) error {
	if sdkProvider, ok := provider.(*sdkmetric.MeterProvider); ok {
		if err := sdkProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("flush metrics: %w", err)
		}
	}
	return nil
}
