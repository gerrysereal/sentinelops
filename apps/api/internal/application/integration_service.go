package application

import (
	"context"
	"fmt"
	"net/url"
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
	if err := validateIntegrationInput(req.Type, req.EndpointURL, req.SyncIntervalSeconds); err != nil {
		return domain.IntegrationConfig{}, err
	}
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
		if err := validateEndpointURL(*req.EndpointURL); err != nil {
			return domain.IntegrationConfig{}, err
		}
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
		if *req.SyncIntervalSeconds < 15 || *req.SyncIntervalSeconds > 86400 {
			return domain.IntegrationConfig{}, fmt.Errorf("sync interval must be between 15 and 86400 seconds")
		}
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

func (s *IntegrationService) Status(ctx context.Context, id string) (domain.IntegrationStatusView, error) {
	return s.repo.GetIntegrationStatus(ctx, id)
}

func (s *IntegrationService) ConnectionHistory(ctx context.Context, id string) ([]domain.ConnectionHistory, error) {
	return s.repo.ListConnectionHistory(ctx, id)
}

func (s *IntegrationService) Types(ctx context.Context) []domain.IntegrationType {
	return []domain.IntegrationType{
		{Type: "Prometheus", Category: "Observability", Description: "Metrics provider health and query integration"},
		{Type: "Grafana", Category: "Observability", Description: "Dashboard and visualization integration"},
		{Type: "Harbor", Category: "Registry", Description: "Container registry integration"},
		{Type: "Vault", Category: "Secrets", Description: "Vault/OpenBao secret management integration"},
		{Type: "GitHub", Category: "SCM", Description: "Repository provider integration"},
		{Type: "GitHubActions", Category: "CI/CD", Description: "Pipeline provider integration"},
		{Type: "ArgoCD", Category: "GitOps", Description: "GitOps deployment integration"},
		{Type: "Kubernetes", Category: "Runtime", Description: "Cluster inventory and workload integration"},
		{Type: "Loki", Category: "Observability", Description: "Log query integration"},
		{Type: "Tempo", Category: "Observability", Description: "Trace query integration"},
		{Type: "SonarQube", Category: "Security", Description: "Code quality and SAST integration"},
		{Type: "Trivy", Category: "Security", Description: "Vulnerability scanner integration"},
		{Type: "Semgrep", Category: "Security", Description: "Static analysis scanner integration"},
		{Type: "Gitleaks", Category: "Security", Description: "Secret scanning integration"},
		{Type: "Falco", Category: "Runtime Security", Description: "Runtime detection integration"},
		{Type: "Wazuh", Category: "SIEM", Description: "SIEM and compliance integration"},
		{Type: "Keycloak", Category: "Authentication", Description: "Identity provider integration"},
		{Type: "OPA-Gatekeeper", Category: "Policy", Description: "Kubernetes policy enforcement integration"},
	}
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

func validateIntegrationInput(toolType string, endpointURL string, syncInterval int) error {
	if strings.TrimSpace(toolType) == "" {
		return fmt.Errorf("integration type is required")
	}
	if err := validateEndpointURL(endpointURL); err != nil {
		return err
	}
	if syncInterval != 0 && (syncInterval < 15 || syncInterval > 86400) {
		return fmt.Errorf("sync interval must be between 15 and 86400 seconds")
	}
	return nil
}

func validateEndpointURL(endpoint string) error {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("endpoint URL must be a valid absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("endpoint URL scheme must be http or https")
	}
	return nil
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
