package observability

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultServiceName = "sentinelops-api"

// Config contains all runtime settings required by the SentinelOps
// observability SDK. It intentionally uses OpenTelemetry-compatible
// environment variable names so the same deployment configuration works in
// Docker, Kubernetes, and future GitOps-managed environments.
type Config struct {
	Enabled               bool
	ServiceName           string
	ServiceVersion        string
	DeploymentEnvironment string
	OTLPEndpoint          string
	OTLPInsecure          bool
	OTLPHeaders           map[string]string
	ExportTimeout         time.Duration
	BatchTimeout          time.Duration
	MetricExportInterval  time.Duration
	Sampler               string
	SamplerArgument       float64
}

// LoadConfigFromEnv loads OpenTelemetry settings from environment variables.
// Standard OpenTelemetry variables are preferred. SentinelOps-specific
// variables are not introduced unless OpenTelemetry has no equivalent.
func LoadConfigFromEnv() Config {
	return Config{
		Enabled:               !envBool("OTEL_SDK_DISABLED", false),
		ServiceName:           envString("OTEL_SERVICE_NAME", DefaultServiceName),
		ServiceVersion:        strings.TrimSpace(os.Getenv("OTEL_SERVICE_VERSION")),
		DeploymentEnvironment: deploymentEnvironment(),
		OTLPEndpoint:          exporterEndpoint(),
		OTLPInsecure:          exporterInsecure(),
		OTLPHeaders:           parseHeaders(firstNonEmpty(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"), os.Getenv("OTEL_EXPORTER_OTLP_TRACES_HEADERS"))),
		ExportTimeout:         envMillisecondsOrDuration("OTEL_EXPORTER_OTLP_TIMEOUT", 10*time.Second),
		BatchTimeout:          envMillisecondsOrDuration("OTEL_BSP_SCHEDULE_DELAY", 5*time.Second),
		MetricExportInterval:  envMillisecondsOrDuration("OTEL_METRIC_EXPORT_INTERVAL", 30*time.Second),
		Sampler:               envString("OTEL_TRACES_SAMPLER", "parentbased_always_on"),
		SamplerArgument:       envFloat("OTEL_TRACES_SAMPLER_ARG", 1),
	}
}

// Validate rejects invalid telemetry configuration before exporters are
// initialized. Disabled telemetry remains valid and does not require an
// endpoint.
func (c Config) Validate() error {
	var issues []string

	if strings.TrimSpace(c.ServiceName) == "" {
		issues = append(issues, "OTEL_SERVICE_NAME is required")
	}
	if c.ExportTimeout <= 0 {
		issues = append(issues, "OTEL_EXPORTER_OTLP_TIMEOUT must be greater than zero")
	}
	if c.BatchTimeout <= 0 {
		issues = append(issues, "OTEL_BSP_SCHEDULE_DELAY must be greater than zero")
	}
	if c.MetricExportInterval <= 0 {
		issues = append(issues, "OTEL_METRIC_EXPORT_INTERVAL must be greater than zero")
	}
	if c.SamplerArgument < 0 || c.SamplerArgument > 1 {
		issues = append(issues, "OTEL_TRACES_SAMPLER_ARG must be between 0 and 1")
	}
	if _, err := samplerFromConfig(c.Sampler, c.SamplerArgument); err != nil {
		issues = append(issues, err.Error())
	}
	if c.Enabled {
		if strings.TrimSpace(c.OTLPEndpoint) == "" {
			issues = append(issues, "OTEL_EXPORTER_OTLP_ENDPOINT is required when telemetry is enabled")
		} else if _, _, err := normalizeGRPCEndpoint(c.OTLPEndpoint, c.OTLPInsecure); err != nil {
			issues = append(issues, err.Error())
		}
	}

	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}

func exporterEndpoint() string {
	return firstNonEmpty(
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	)
}

func exporterInsecure() bool {
	if raw, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_INSECURE"); ok {
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err == nil {
			return parsed
		}
	}

	rawEndpoint := firstNonEmpty(
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	)
	parsed, err := url.Parse(rawEndpoint)
	if err == nil && parsed.Scheme != "" {
		return strings.EqualFold(parsed.Scheme, "http")
	}

	// The in-cluster collector endpoint is plaintext gRPC by default. A TLS
	// endpoint can override this with OTEL_EXPORTER_OTLP_INSECURE=false.
	return true
}

func deploymentEnvironment() string {
	return firstNonEmpty(
		os.Getenv("OTEL_DEPLOYMENT_ENVIRONMENT"),
		os.Getenv("APP_ENV"),
	)
}

func normalizeGRPCEndpoint(raw string, insecure bool) (string, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", insecure, errors.New("OTLP gRPC endpoint is empty")
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", insecure, fmt.Errorf("parse OTLP gRPC endpoint: %w", err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "", insecure, fmt.Errorf("OTLP gRPC endpoint scheme must be http or https")
		}
		if parsed.Host == "" {
			return "", insecure, errors.New("OTLP gRPC endpoint host is required")
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return "", insecure, errors.New("OTLP gRPC endpoint must not contain a path")
		}
		value = parsed.Host
		insecure = parsed.Scheme == "http"
	}

	if _, _, err := net.SplitHostPort(value); err != nil {
		return "", insecure, fmt.Errorf("OTLP gRPC endpoint must include host and port: %w", err)
	}
	return value, insecure, nil
}

func parseHeaders(raw string) map[string]string {
	result := make(map[string]string)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, err := url.QueryUnescape(strings.TrimSpace(parts[0]))
		if err != nil || key == "" {
			continue
		}
		value, err := url.QueryUnescape(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		result[key] = value
	}
	return result
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envMillisecondsOrDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if milliseconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Duration(milliseconds) * time.Millisecond
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
