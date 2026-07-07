package application

import (
	"context"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
	"github.com/sentinelops/sentinelops/apps/api/internal/platform/database"
)

type ModuleService struct {
	repo *database.Repository
}

func NewModuleService(repo *database.Repository) *ModuleService {
	return &ModuleService{repo: repo}
}

func (s *ModuleService) ListApplications(ctx context.Context) ([]domain.Application, error) {
	return s.repo.ListApplications(ctx)
}

func (s *ModuleService) CreateApplication(ctx context.Context, req domain.CreateApplicationRequest) (domain.Application, error) {
	return s.repo.CreateApplication(ctx, req)
}

func (s *ModuleService) ListPipelines(ctx context.Context) ([]domain.PipelineRun, error) {
	return s.repo.ListPipelines(ctx)
}

func (s *ModuleService) ListDeployments(ctx context.Context) ([]domain.Deployment, error) {
	return s.repo.ListDeployments(ctx)
}

func (s *ModuleService) ListSecurityAlerts(ctx context.Context) ([]domain.SecurityAlert, error) {
	return s.repo.ListSecurityAlerts(ctx, 100)
}

func (s *ModuleService) ListObservabilitySignals(ctx context.Context) ([]domain.ObservabilitySignal, error) {
	return s.repo.ListObservabilitySignals(ctx)
}

func (s *ModuleService) ListIntegrations(ctx context.Context) ([]domain.Integration, error) {
	return defaultIntegrations(), nil
}
