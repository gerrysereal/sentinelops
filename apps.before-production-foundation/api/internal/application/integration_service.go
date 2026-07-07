package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	secretcrypto "github.com/sentinelops/sentinelops/apps/api/internal/platform/crypto"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
	"github.com/sentinelops/sentinelops/apps/api/internal/ports"
)

type IntegrationService struct {
	repo      *database.Repository
	secretBox *secretcrypto.SecretBox
	registry  ports.IntegrationRegistry
	mode      string
}

func NewIntegrationService(repo *database.Repository, secretBox *secretcrypto.SecretBox, registry ports.IntegrationRegistry, mode string) *IntegrationService {
	if strings.TrimSpace(mode) == "" {
		mode = domain.IntegrationModeDemo
	}
	return &IntegrationService{repo: repo, secretBox: secretBox, registry: registry, mode: mode}
}

func (s *IntegrationService) List(ctx context.Context) ([]domain.IntegrationConfig, error) {
	return s.repo.ListIntegrations(ctx)
}

func (s *IntegrationService) Get(ctx context.Context, id string) (domain.IntegrationConfig, error) {
	item, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	return s.decrypt(item)
}

func (s *IntegrationService) Create(ctx context.Context, req domain.CreateIntegrationRequest) (domain.IntegrationConfig, error) {
	cfg := domain.IntegrationConfig{
		Name:                strings.TrimSpace(req.Name),
		Type:                strings.TrimSpace(req.Type),
		Category:            categoryFor(req.Type, req.Category),
		EndpointURL:         strings.TrimSpace(req.EndpointURL),
		Username:            strings.TrimSpace(req.Username),
		Namespace:           strings.TrimSpace(req.Namespace),
		TLSVerify:           true,
		SyncIntervalSeconds: req.SyncIntervalSeconds,
		Enabled:             req.Enabled,
		Status:              domain.IntegrationStatusDisabled,
		Mode:                s.mode,
		Health:              "Not tested yet",
	}
	if req.TLSVerify != nil {
		cfg.TLSVerify = *req.TLSVerify
	}
	if cfg.SyncIntervalSeconds == 0 {
		cfg.SyncIntervalSeconds = 60
	}
	if cfg.Enabled {
		cfg.Status = domain.IntegrationStatusUnknown
	}

	var err error
	cfg.AccessToken, err = s.secretBox.Encrypt(req.AccessToken)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	cfg.Password, err = s.secretBox.Encrypt(req.Password)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}

	return s.repo.CreateIntegration(ctx, cfg)
}

func (s *IntegrationService) Update(ctx context.Context, id string, req domain.UpdateIntegrationRequest) (domain.IntegrationConfig, error) {
	cfg, err := s.repo.GetIntegration(ctx, id)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	if req.Name != nil {
		cfg.Name = strings.TrimSpace(*req.Name)
	}
	if req.Category != nil {
		cfg.Category = categoryFor(cfg.Type, *req.Category)
	}
	if req.EndpointURL != nil {
		cfg.EndpointURL = strings.TrimSpace(*req.EndpointURL)
	}
	if req.Username != nil {
		cfg.Username = strings.TrimSpace(*req.Username)
	}
	if req.Namespace != nil {
		cfg.Namespace = strings.TrimSpace(*req.Namespace)
	}
	if req.TLSVerify != nil {
		cfg.TLSVerify = *req.TLSVerify
	}
	if req.SyncIntervalSeconds != nil {
		cfg.SyncIntervalSeconds = *req.SyncIntervalSeconds
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
		if !cfg.Enabled {
			cfg.Status = domain.IntegrationStatusDisabled
		} else if cfg.Status == domain.IntegrationStatusDisabled {
			cfg.Status = domain.IntegrationStatusUnknown
		}
	}
	if req.AccessToken != nil {
		cfg.AccessToken, err = s.secretBox.Encrypt(*req.AccessToken)
		if err != nil {
			return domain.IntegrationConfig{}, err
		}
	}
	if req.Password != nil {
		cfg.Password, err = s.secretBox.Encrypt(*req.Password)
		if err != nil {
			return domain.IntegrationConfig{}, err
		}
	}
	return s.repo.UpdateIntegration(ctx, cfg)
}

func (s *IntegrationService) Delete(ctx context.Context, id string) error {
	return s.repo.DeleteIntegration(ctx, id)
}

func (s *IntegrationService) SetEnabled(ctx context.Context, id string, req domain.SetIntegrationEnabledRequest) (domain.IntegrationConfig, error) {
	return s.repo.SetIntegrationEnabled(ctx, id, req.Enabled)
}

func (s *IntegrationService) TestConnection(ctx context.Context, id string) (domain.IntegrationHealth, error) {
	cfg, err := s.Get(ctx, id)
	if err != nil {
		return domain.IntegrationHealth{}, err
	}
	health, err := s.registry.For(ctx, cfg).Health(ctx, cfg)
	if err != nil {
		return domain.IntegrationHealth{}, err
	}
	_, err = s.repo.UpdateIntegrationHealth(ctx, id, health)
	return health, err
}

func (s *IntegrationService) Sync(ctx context.Context, id string) (domain.IntegrationSyncResult, error) {
	cfg, err := s.Get(ctx, id)
	if err != nil {
		return domain.IntegrationSyncResult{}, err
	}
	result, err := s.registry.For(ctx, cfg).Sync(ctx, cfg)
	if err != nil {
		return domain.IntegrationSyncResult{}, err
	}
	_, err = s.repo.MarkIntegrationSynced(ctx, id, result)
	if err != nil {
		return domain.IntegrationSyncResult{}, err
	}
	_ = s.repo.CreateObservabilitySignal(ctx, cfg.Type, "integration_sync", statusToSignal(result.Status), fmt.Sprintf("%s: %s", cfg.Name, result.Message))
	return result, nil
}

func (s *IntegrationService) Logs(ctx context.Context, id string) ([]domain.IntegrationLog, error) {
	return s.repo.ListIntegrationLogs(ctx, id)
}

func (s *IntegrationService) decrypt(cfg domain.IntegrationConfig) (domain.IntegrationConfig, error) {
	var err error
	cfg.AccessToken, err = s.secretBox.Decrypt(cfg.AccessToken)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	cfg.Password, err = s.secretBox.Decrypt(cfg.Password)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	return cfg, nil
}

func categoryFor(toolType string, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		return fallback
	}
	switch toolType {
	case "Prometheus", "Grafana", "Loki", "Tempo":
		return "Observability"
	case "Harbor":
		return "Registry"
	case "Vault":
		return "Secrets"
	case "GitHub", "GitHubActions":
		return "CI/CD"
	case "ArgoCD", "Kubernetes":
		return "GitOps"
	case "SonarQube", "Trivy", "Semgrep", "Gitleaks", "Falco", "Wazuh", "OPA-Gatekeeper":
		return "Security"
	case "Keycloak":
		return "Authentication"
	default:
		return "Platform"
	}
}

func statusToSignal(status string) string {
	switch status {
	case domain.IntegrationStatusConnected:
		return "healthy"
	case domain.IntegrationStatusDisabled:
		return "warning"
	default:
		return "warning"
	}
}
