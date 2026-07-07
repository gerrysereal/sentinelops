package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
)

type OverviewService struct {
	repo  *database.Repository
	redis *redis.Client
}

func NewOverviewService(repo *database.Repository, redisClient *redis.Client) *OverviewService {
	return &OverviewService{repo: repo, redis: redisClient}
}

func (s *OverviewService) GetOverview(ctx context.Context) (domain.Overview, error) {
	const cacheKey = "sentinelops:overview:v1"

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var overview domain.Overview
			if json.Unmarshal([]byte(cached), &overview) == nil {
				return overview, nil
			}
		}
	}

	applicationsCount, err := s.repo.CountApplications(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	deploymentStatus, err := s.repo.CountDeploymentsByHealth(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	pipelineStatus, err := s.repo.CountPipelinesByStatus(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	securitySeverity, err := s.repo.CountSecurityBySeverity(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	recentAlerts, err := s.repo.ListSecurityAlerts(ctx, 5)
	if err != nil {
		return domain.Overview{}, err
	}

	overview := domain.Overview{
		ApplicationsCount: applicationsCount,
		ClustersCount:     2,
		NodesCount:        8,
		PodsCount:         52,
		DeploymentStatus:  deploymentStatus,
		PipelineStatus:    pipelineStatus,
		SecuritySeverity:  securitySeverity,
		ResourceUsage: map[string]int{
			"cpu":     43,
			"memory":  61,
			"storage": 39,
		},
		RecentAlerts: recentAlerts,
		Integrations: defaultIntegrations(),
	}

	if s.redis != nil {
		payload, _ := json.Marshal(overview)
		_ = s.redis.Set(ctx, cacheKey, payload, 30*time.Second).Err()
	}

	return overview, nil
}

func defaultIntegrations() []domain.Integration {
	return []domain.Integration{
		{Name: "Argo CD", Category: "GitOps", Status: "healthy", Endpoint: "argocd.sentinelops.local"},
		{Name: "Harbor", Category: "Registry", Status: "healthy", Endpoint: "harbor.sentinelops.local"},
		{Name: "Vault/OpenBao", Category: "Secrets", Status: "healthy", Endpoint: "vault.sentinelops.local"},
		{Name: "Prometheus", Category: "Metrics", Status: "healthy", Endpoint: "prometheus.sentinelops.local"},
		{Name: "Grafana", Category: "Visualization", Status: "healthy", Endpoint: "grafana.sentinelops.local"},
		{Name: "Loki", Category: "Logging", Status: "healthy", Endpoint: "loki.sentinelops.local"},
		{Name: "Tempo", Category: "Tracing", Status: "degraded", Endpoint: "tempo.sentinelops.local"},
		{Name: "Falco", Category: "Runtime Security", Status: "warning", Endpoint: "falco.sentinelops.local"},
		{Name: "Wazuh", Category: "SIEM", Status: "healthy", Endpoint: "wazuh.sentinelops.local"},
		{Name: "OPA Gatekeeper", Category: "Policy", Status: "healthy", Endpoint: "gatekeeper.sentinelops.local"},
	}
}
