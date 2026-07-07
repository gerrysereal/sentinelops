package application

import (
	"context"
	"fmt"

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

func (s *ModuleService) RunPipeline(ctx context.Context, req domain.CreatePipelineRunRequest) (domain.PipelineRun, error) {
	return domain.PipelineRun{}, fmt.Errorf("pipeline execution is disabled until a real GitHubActions adapter is connected")
}

func (s *ModuleService) ListDeployments(ctx context.Context) ([]domain.Deployment, error) {
	return s.repo.ListDeployments(ctx)
}

func (s *ModuleService) CreateDeployment(ctx context.Context, req domain.CreateDeploymentRequest) (domain.Deployment, error) {
	return domain.Deployment{}, fmt.Errorf("deployment execution is disabled until a real ArgoCD or Kubernetes adapter is connected")
}

func (s *ModuleService) ListSecurityAlerts(ctx context.Context) ([]domain.SecurityAlert, error) {
	return s.repo.ListSecurityAlerts(ctx, 100)
}

func (s *ModuleService) RunSecurityScan(ctx context.Context, req domain.CreateSecurityScanRequest) (domain.SecurityAlert, error) {
	return domain.SecurityAlert{}, fmt.Errorf("security scanning is disabled until a real scanner adapter is connected")
}

func (s *ModuleService) UpdateSecurityAlertStatus(ctx context.Context, id string, req domain.UpdateAlertStatusRequest) (domain.SecurityAlert, error) {
	return s.repo.UpdateSecurityAlertStatus(ctx, id, req.Status)
}

func (s *ModuleService) ListObservabilitySignals(ctx context.Context) ([]domain.ObservabilitySignal, error) {
	return s.repo.ListObservabilitySignals(ctx)
}

func (s *ModuleService) ListRegistryArtifacts(ctx context.Context) ([]domain.RegistryArtifact, error) {
	return s.repo.ListRegistryArtifacts(ctx)
}

func (s *ModuleService) ListSettings(ctx context.Context) ([]domain.Setting, error) {
	return s.repo.ListSettings(ctx)
}

func (s *ModuleService) UpsertSetting(ctx context.Context, key string, req domain.UpdateSettingRequest) (domain.Setting, error) {
	return s.repo.UpsertSetting(ctx, key, req)
}
