package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sentinelops/sentinelops/apps/api/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListApplications(ctx context.Context) ([]domain.Application, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, owner_name, repository, environment, status, created_at
		FROM applications
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Application, 0)
	for rows.Next() {
		var item domain.Application
		if err := rows.Scan(&item.ID, &item.Name, &item.Owner, &item.Repository, &item.Environment, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateApplication(ctx context.Context, req domain.CreateApplicationRequest) (domain.Application, error) {
	item := domain.Application{
		ID:          "app-" + randomHex(8),
		Name:        req.Name,
		Owner:       req.Owner,
		Repository:  req.Repository,
		Environment: req.Environment,
		Status:      "healthy",
	}

	if err := r.pool.QueryRow(ctx, `
		INSERT INTO applications (id, name, owner_name, repository, environment, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`, item.ID, item.Name, item.Owner, item.Repository, item.Environment, item.Status).Scan(&item.CreatedAt); err != nil {
		return domain.Application{}, err
	}

	return item, nil
}

func (r *Repository) ListPipelines(ctx context.Context) ([]domain.PipelineRun, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.application_id, a.name, p.branch, p.commit_sha, p.status, p.stage, p.duration_seconds, p.finished_at
		FROM pipeline_runs p
		JOIN applications a ON a.id = p.application_id
		ORDER BY p.finished_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PipelineRun, 0)
	for rows.Next() {
		var item domain.PipelineRun
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.ApplicationName, &item.Branch, &item.CommitSHA, &item.Status, &item.Stage, &item.DurationSeconds, &item.FinishedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListDeployments(ctx context.Context) ([]domain.Deployment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.application_id, a.name, d.cluster_name, d.namespace_name, d.image, d.version, d.sync_status, d.health_status, d.deployed_at
		FROM deployments d
		JOIN applications a ON a.id = d.application_id
		ORDER BY d.deployed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Deployment, 0)
	for rows.Next() {
		var item domain.Deployment
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.ApplicationName, &item.Cluster, &item.Namespace, &item.Image, &item.Version, &item.SyncStatus, &item.HealthStatus, &item.DeployedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListSecurityAlerts(ctx context.Context, limit int) ([]domain.SecurityAlert, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source, severity, title, application_name, status, detected_at
		FROM security_alerts
		ORDER BY detected_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.SecurityAlert, 0)
	for rows.Next() {
		var item domain.SecurityAlert
		if err := rows.Scan(&item.ID, &item.Source, &item.Severity, &item.Title, &item.Application, &item.Status, &item.DetectedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ListObservabilitySignals(ctx context.Context) ([]domain.ObservabilitySignal, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, source, signal_type, status, message, created_at
		FROM observability_signals
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ObservabilitySignal, 0)
	for rows.Next() {
		var item domain.ObservabilitySignal
		if err := rows.Scan(&item.ID, &item.Source, &item.Type, &item.Status, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CountApplications(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications`).Scan(&count)
	return count, err
}

func (r *Repository) CountDeploymentsByHealth(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT health_status, COUNT(*) FROM deployments GROUP BY health_status`)
}

func (r *Repository) CountPipelinesByStatus(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT status, COUNT(*) FROM pipeline_runs GROUP BY status`)
}

func (r *Repository) CountSecurityBySeverity(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT severity, COUNT(*) FROM security_alerts GROUP BY severity`)
}

func (r *Repository) countBy(ctx context.Context, query string) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var key string
		var value int
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, rows.Err()
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("fallback-%d", size)
	}
	return hex.EncodeToString(buf)
}
