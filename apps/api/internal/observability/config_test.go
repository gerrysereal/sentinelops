package observability

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "false")
	t.Setenv("OTEL_SERVICE_NAME", "sentinelops-api-test")
	t.Setenv("OTEL_SERVICE_VERSION", "1.2.3")
	t.Setenv("OTEL_DEPLOYMENT_ENVIRONMENT", "test")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector.test:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "authorization=Bearer%20token,x-tenant=sentinelops")
	t.Setenv("OTEL_EXPORTER_OTLP_TIMEOUT", "2500")
	t.Setenv("OTEL_BSP_SCHEDULE_DELAY", "750ms")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "15000")
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "0.25")

	cfg := LoadConfigFromEnv()

	if !cfg.Enabled {
		t.Fatal("expected telemetry to be enabled")
	}
	if cfg.ServiceName != "sentinelops-api-test" {
		t.Fatalf("unexpected service name: %s", cfg.ServiceName)
	}
	if cfg.OTLPEndpoint != "http://collector.test:4317" {
		t.Fatalf("unexpected endpoint: %s", cfg.OTLPEndpoint)
	}
	if !cfg.OTLPInsecure {
		t.Fatal("expected HTTP endpoint to imply insecure gRPC")
	}
	if cfg.OTLPHeaders["authorization"] != "Bearer token" {
		t.Fatalf("unexpected decoded authorization header: %q", cfg.OTLPHeaders["authorization"])
	}
	if cfg.ExportTimeout != 2500*time.Millisecond {
		t.Fatalf("unexpected export timeout: %s", cfg.ExportTimeout)
	}
	if cfg.BatchTimeout != 750*time.Millisecond {
		t.Fatalf("unexpected batch timeout: %s", cfg.BatchTimeout)
	}
	if cfg.MetricExportInterval != 15*time.Second {
		t.Fatalf("unexpected metric interval: %s", cfg.MetricExportInterval)
	}
	if cfg.SamplerArgument != 0.25 {
		t.Fatalf("unexpected sampler argument: %f", cfg.SamplerArgument)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to be valid: %v", err)
	}
}

func TestConfigValidateRejectsInvalidSampler(t *testing.T) {
	cfg := Config{
		Enabled:              true,
		ServiceName:          DefaultServiceName,
		OTLPEndpoint:         "otel-opentelemetry-collector.monitoring.svc.cluster.local:4317",
		OTLPInsecure:         true,
		ExportTimeout:        time.Second,
		BatchTimeout:         time.Second,
		MetricExportInterval: time.Second,
		Sampler:              "unsupported",
		SamplerArgument:      1,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unsupported sampler to fail validation")
	}
}

func TestNormalizeGRPCEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		insecure         bool
		wantEndpoint     string
		wantInsecure     bool
		expectValidation bool
	}{
		{
			name:         "plain host and port",
			input:        "collector.monitoring.svc:4317",
			insecure:     true,
			wantEndpoint: "collector.monitoring.svc:4317",
			wantInsecure: true,
		},
		{
			name:         "http url",
			input:        "http://collector.monitoring.svc:4317",
			wantEndpoint: "collector.monitoring.svc:4317",
			wantInsecure: true,
		},
		{
			name:         "https url",
			input:        "https://collector.example.com:4317",
			insecure:     true,
			wantEndpoint: "collector.example.com:4317",
			wantInsecure: false,
		},
		{
			name:             "path rejected",
			input:            "https://collector.example.com:4317/v1/traces",
			expectValidation: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint, insecure, err := normalizeGRPCEndpoint(test.input, test.insecure)
			if test.expectValidation {
				if err == nil {
					t.Fatal("expected validation error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if endpoint != test.wantEndpoint {
				t.Fatalf("unexpected endpoint: got %q want %q", endpoint, test.wantEndpoint)
			}
			if insecure != test.wantInsecure {
				t.Fatalf("unexpected insecure flag: got %t want %t", insecure, test.wantInsecure)
			}
		})
	}
}
