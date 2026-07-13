package observability

import (
	"context"
	"testing"
	"time"
)

func TestDisabledSDKDoesNotOpenExporterConnection(t *testing.T) {
	cfg := Config{
		Enabled:              false,
		ServiceName:          DefaultServiceName,
		ExportTimeout:        time.Second,
		BatchTimeout:         time.Second,
		MetricExportInterval: time.Second,
		Sampler:              "parentbased_always_on",
		SamplerArgument:      1,
	}

	sdk, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("initialize disabled SDK: %v", err)
	}
	if sdk.TracerProvider() == nil {
		t.Fatal("expected a no-op tracer provider")
	}
	if sdk.MeterProvider() == nil {
		t.Fatal("expected a no-op meter provider")
	}
	if sdk.Propagator() == nil {
		t.Fatal("expected a propagator")
	}
	if err := sdk.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown disabled SDK: %v", err)
	}
}
