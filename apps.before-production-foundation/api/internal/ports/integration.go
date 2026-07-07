package ports

import (
	"context"

	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
)

type Integration interface {
	Connect(ctx context.Context, cfg domain.IntegrationConfig) error
	Health(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationHealth, error)
	Sync(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationSyncResult, error)
	Status(ctx context.Context, cfg domain.IntegrationConfig) (string, error)
	Configuration(ctx context.Context, cfg domain.IntegrationConfig) (map[string]string, error)
}

type IntegrationRegistry interface {
	For(ctx context.Context, cfg domain.IntegrationConfig) Integration
}
