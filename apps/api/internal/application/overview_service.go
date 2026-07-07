package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
)

const overviewCacheKey = "sentinelops:overview:v1"

type OverviewService struct {
	repo  *database.Repository
	redis *redis.Client
}

func NewOverviewService(repo *database.Repository, redisClient *redis.Client) *OverviewService {
	return &OverviewService{repo: repo, redis: redisClient}
}

func (s *OverviewService) Invalidate(ctx context.Context) {
	if s.redis != nil {
		_ = s.redis.Del(ctx, overviewCacheKey).Err()
	}
}

func (s *OverviewService) GetOverview(ctx context.Context) (domain.Overview, error) {
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, overviewCacheKey).Result()
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

	clustersCount, err := s.repo.CountClusters(ctx)
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

	integrationList, err := s.repo.ListIntegrations(ctx)
	if err != nil {
		return domain.Overview{}, err
	}

	overview := domain.Overview{
		ApplicationsCount: applicationsCount,
		ClustersCount:     clustersCount,
		NodesCount:        0,
		PodsCount:         0,
		DeploymentStatus:  deploymentStatus,
		PipelineStatus:    pipelineStatus,
		SecuritySeverity:  securitySeverity,
		ResourceUsage: map[string]int{
			"cpu":     0,
			"memory":  0,
			"storage": 0,
		},
		RecentAlerts: recentAlerts,
		Integrations: integrationList,
	}

	if s.redis != nil {
		payload, _ := json.Marshal(overview)
		_ = s.redis.Set(ctx, overviewCacheKey, payload, 5*time.Second).Err()
	}

	return overview, nil
}
