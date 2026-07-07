package integrations

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/retry"
	"github.com/sentinelops/sentinelops/apps/api/internal/ports"
)

type RegistryOptions struct {
	Mode    string
	Timeout time.Duration
	Retries int
}

type Registry struct {
	mode    string
	timeout time.Duration
	retries int
}

func NewRegistry(options RegistryOptions) *Registry {
	mode := strings.ToLower(strings.TrimSpace(options.Mode))
	if mode == "" {
		mode = domain.IntegrationModeDemo
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	retries := options.Retries
	if retries < 0 {
		retries = 0
	}
	return &Registry{mode: mode, timeout: timeout, retries: retries}
}

func (r *Registry) For(ctx context.Context, cfg domain.IntegrationConfig) ports.Integration {
	return HTTPAdapter{
		client:  httpClient(cfg.TLSVerify, r.timeout),
		retries: r.retries,
	}
}

type HTTPAdapter struct {
	client  *http.Client
	retries int
}

func (a HTTPAdapter) Connect(ctx context.Context, cfg domain.IntegrationConfig) error {
	result, err := a.Health(ctx, cfg)
	if err != nil {
		return err
	}
	if !result.Healthy {
		return errors.New(result.Message)
	}
	return nil
}

func (a HTTPAdapter) Health(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationHealth, error) {
	started := time.Now().UTC()
	if !cfg.Enabled {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusDisabled,
			Healthy:   false,
			Message:   "integration is disabled",
			CheckedAt: started,
			LatencyMs: 0,
			Attributes: map[string]string{
				"adapter": "http",
				"mode":    cfg.Mode,
			},
		}, nil
	}

	endpoint, err := validatedHealthURL(cfg)
	if err != nil {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusError,
			Healthy:   false,
			Message:   err.Error(),
			CheckedAt: started,
			LatencyMs: int(time.Since(started).Milliseconds()),
			Attributes: map[string]string{
				"adapter": "http",
			},
		}, nil
	}

	var responseStatus int
	var responseBody string
	var lastErr error
	policy := retry.Policy{Attempts: a.retries + 1, Backoff: 250 * time.Millisecond}
	err = policy.Do(ctx, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		applyAuth(req, cfg)
		resp, err := a.client.Do(req)
		if err != nil {
			lastErr = err
			return err
		}
		defer resp.Body.Close()
		responseStatus = resp.StatusCode
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		responseBody = strings.TrimSpace(string(body))
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("remote service returned HTTP %d", resp.StatusCode)
			return lastErr
		}
		return nil
	})

	latency := int(time.Since(started).Milliseconds())
	if err != nil {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusDisconnected,
			Healthy:   false,
			Message:   lastErrOr(err, lastErr),
			CheckedAt: started,
			LatencyMs: latency,
			Attributes: map[string]string{
				"adapter": "http",
				"url":     endpoint,
			},
		}, nil
	}

	healthy := responseStatus >= 200 && responseStatus < 400
	status := domain.IntegrationStatusConnected
	if !healthy {
		status = domain.IntegrationStatusError
	}
	message := fmt.Sprintf("%s responded with HTTP %d", endpoint, responseStatus)
	if !healthy && responseBody != "" {
		message = fmt.Sprintf("%s; body: %.180s", message, responseBody)
	}
	return domain.IntegrationHealth{
		Status:    status,
		Healthy:   healthy,
		Message:   message,
		CheckedAt: started,
		LatencyMs: latency,
		Attributes: map[string]string{
			"statusCode": fmt.Sprintf("%d", responseStatus),
			"adapter":    "http",
			"url":        endpoint,
			"mode":       cfg.Mode,
		},
	}, nil
}

func (a HTTPAdapter) Sync(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationSyncResult, error) {
	health, err := a.Health(ctx, cfg)
	if err != nil {
		return domain.IntegrationSyncResult{}, err
	}
	return domain.IntegrationSyncResult{
		Status:   health.Status,
		Message:  "integration status synchronized from health endpoint",
		SyncedAt: time.Now().UTC(),
		Resources: map[string]int{
			"health_checks": 1,
		},
		Metadata: health.Attributes,
	}, nil
}

func (a HTTPAdapter) Status(ctx context.Context, cfg domain.IntegrationConfig) (string, error) {
	health, err := a.Health(ctx, cfg)
	if err != nil {
		return domain.IntegrationStatusError, err
	}
	return health.Status, nil
}

func (a HTTPAdapter) Configuration(ctx context.Context, cfg domain.IntegrationConfig) (map[string]string, error) {
	endpoint, err := validatedHealthURL(cfg)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"endpointUrl": cfg.EndpointURL,
		"healthUrl":   endpoint,
		"namespace":   cfg.Namespace,
		"tlsVerify":   fmt.Sprintf("%t", cfg.TLSVerify),
		"adapter":     "http",
		"mode":        cfg.Mode,
	}, nil
}

func httpClient(tlsVerify bool, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsVerify}, //nolint:gosec -- explicit integration setting stored per integration.
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

func applyAuth(req *http.Request, cfg domain.IntegrationConfig) {
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	}
	if cfg.Username != "" && cfg.Password != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}
}

func validatedHealthURL(cfg domain.IntegrationConfig) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.EndpointURL), "/")
	if base == "" {
		return "", errors.New("endpoint URL is required")
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("endpoint URL is invalid")
	}
	path := healthPath(cfg.Type)
	if path == "" {
		return base, nil
	}
	return base + path, nil
}

func healthPath(toolType string) string {
	switch toolType {
	case "Prometheus":
		return "/-/healthy"
	case "Grafana":
		return "/api/health"
	case "Harbor":
		return "/api/v2.0/health"
	case "ArgoCD":
		return "/api/version"
	case "Loki":
		return "/ready"
	case "Tempo":
		return "/ready"
	case "Vault":
		return "/v1/sys/health"
	case "SonarQube":
		return "/api/system/health"
	case "Keycloak":
		return "/realms/master"
	default:
		return ""
	}
}

func lastErrOr(primary error, secondary error) string {
	if secondary != nil {
		return secondary.Error()
	}
	if primary != nil {
		return primary.Error()
	}
	return "connection failed"
}
