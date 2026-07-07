package integrations

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/ports"
)

type Registry struct {
	mode string
}

func NewRegistry(mode string) *Registry {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = domain.IntegrationModeDemo
	}
	return &Registry{mode: mode}
}

func (r *Registry) For(ctx context.Context, cfg domain.IntegrationConfig) ports.Integration {
	if r.mode == domain.IntegrationModeDemo || cfg.Mode == domain.IntegrationModeDemo {
		return SimulatorAdapter{}
	}
	return HTTPAdapter{client: httpClient(cfg.TLSVerify)}
}

type SimulatorAdapter struct{}

func (SimulatorAdapter) Connect(ctx context.Context, cfg domain.IntegrationConfig) error {
	if !cfg.Enabled {
		return fmt.Errorf("integration is disabled")
	}
	return nil
}

func (SimulatorAdapter) Health(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationHealth, error) {
	if !cfg.Enabled {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusDisabled,
			Healthy:   false,
			Message:   "integration is disabled",
			CheckedAt: time.Now().UTC(),
		}, nil
	}
	return domain.IntegrationHealth{
		Status:    domain.IntegrationStatusConnected,
		Healthy:   true,
		Message:   fmt.Sprintf("%s simulator adapter is reachable in demo mode", cfg.Type),
		CheckedAt: time.Now().UTC(),
		Attributes: map[string]string{
			"mode":      domain.IntegrationModeDemo,
			"namespace": cfg.Namespace,
		},
	}, nil
}

func (SimulatorAdapter) Sync(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationSyncResult, error) {
	if !cfg.Enabled {
		return domain.IntegrationSyncResult{
			Status:   domain.IntegrationStatusDisabled,
			Message:  "integration is disabled",
			SyncedAt: time.Now().UTC(),
		}, nil
	}
	return domain.IntegrationSyncResult{
		Status:   domain.IntegrationStatusConnected,
		Message:  fmt.Sprintf("%s simulator sync completed", cfg.Type),
		SyncedAt: time.Now().UTC(),
		Resources: map[string]int{
			"metrics":  3,
			"events":   1,
			"findings": 0,
		},
		Metadata: map[string]string{
			"adapter": "simulator",
			"mode":    domain.IntegrationModeDemo,
		},
	}, nil
}

func (SimulatorAdapter) Status(ctx context.Context, cfg domain.IntegrationConfig) (string, error) {
	if !cfg.Enabled {
		return domain.IntegrationStatusDisabled, nil
	}
	return domain.IntegrationStatusConnected, nil
}

func (SimulatorAdapter) Configuration(ctx context.Context, cfg domain.IntegrationConfig) (map[string]string, error) {
	return map[string]string{
		"endpointUrl": cfg.EndpointURL,
		"namespace":   cfg.Namespace,
		"tlsVerify":   fmt.Sprintf("%t", cfg.TLSVerify),
		"mode":        domain.IntegrationModeDemo,
	}, nil
}

type HTTPAdapter struct {
	client *http.Client
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
	if !cfg.Enabled {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusDisabled,
			Healthy:   false,
			Message:   "integration is disabled",
			CheckedAt: time.Now().UTC(),
		}, nil
	}

	url := healthURL(cfg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.IntegrationHealth{}, err
	}
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	}
	if cfg.Username != "" && cfg.Password != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return domain.IntegrationHealth{
			Status:    domain.IntegrationStatusDisconnected,
			Healthy:   false,
			Message:   err.Error(),
			CheckedAt: time.Now().UTC(),
		}, nil
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 400
	status := domain.IntegrationStatusConnected
	if !healthy {
		status = domain.IntegrationStatusError
	}
	return domain.IntegrationHealth{
		Status:    status,
		Healthy:   healthy,
		Message:   fmt.Sprintf("%s responded with HTTP %d", url, resp.StatusCode),
		CheckedAt: time.Now().UTC(),
		Attributes: map[string]string{
			"statusCode": fmt.Sprintf("%d", resp.StatusCode),
			"adapter":    "http",
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
		Message:  "health data synchronized from integration endpoint",
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
	return map[string]string{
		"endpointUrl": cfg.EndpointURL,
		"namespace":   cfg.Namespace,
		"tlsVerify":   fmt.Sprintf("%t", cfg.TLSVerify),
		"adapter":     "http",
	}, nil
}

func httpClient(tlsVerify bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsVerify}, //nolint:gosec -- user-controlled lab setting.
	}
	return &http.Client{Timeout: 8 * time.Second, Transport: transport}
}

func healthURL(cfg domain.IntegrationConfig) string {
	base := strings.TrimRight(cfg.EndpointURL, "/")
	switch cfg.Type {
	case "Prometheus":
		return base + "/-/healthy"
	case "Grafana":
		return base + "/api/health"
	case "Harbor":
		return base + "/api/v2.0/health"
	case "ArgoCD":
		return base + "/api/version"
	case "Loki":
		return base + "/ready"
	case "Tempo":
		return base + "/ready"
	case "Vault":
		return base + "/v1/sys/health"
	case "SonarQube":
		return base + "/api/system/health"
	case "Keycloak":
		return base + "/realms/master"
	default:
		return base
	}
}
