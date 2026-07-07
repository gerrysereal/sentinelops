package database

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

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
		Name:        normalizeName(req.Name),
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

func (r *Repository) CreatePipelineRun(ctx context.Context, req domain.CreatePipelineRunRequest) (domain.PipelineRun, error) {
	branch := valueOr(req.Branch, "main")
	stage := valueOr(req.Stage, "security-scan")
	status := valueOr(req.Status, "success")
	duration := 80 + int(randomByte()%220)
	if status == "pending" || status == "running" {
		duration = 0
	}

	item := domain.PipelineRun{
		ID:              "pipe-" + randomHex(8),
		ApplicationID:   req.ApplicationID,
		Branch:          branch,
		CommitSHA:       randomHex(4),
		Status:          status,
		Stage:           stage,
		DurationSeconds: duration,
	}

	if err := r.pool.QueryRow(ctx, `
		INSERT INTO pipeline_runs (id, application_id, branch, commit_sha, status, stage, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING finished_at`, item.ID, item.ApplicationID, item.Branch, item.CommitSHA, item.Status, item.Stage, item.DurationSeconds).Scan(&item.FinishedAt); err != nil {
		return domain.PipelineRun{}, err
	}

	if err := r.pool.QueryRow(ctx, `SELECT name FROM applications WHERE id = $1`, item.ApplicationID).Scan(&item.ApplicationName); err != nil {
		return domain.PipelineRun{}, err
	}

	_, _ = r.pool.Exec(ctx, `
		INSERT INTO observability_signals (id, source, signal_type, status, message)
		VALUES ($1, 'GitHub Actions', 'event', $2, $3)`, "obs-"+randomHex(8), statusToSignal(status), fmt.Sprintf("Pipeline %s for %s finished with status %s", item.ID, item.ApplicationName, status))

	return item, nil
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

func (r *Repository) CreateDeployment(ctx context.Context, req domain.CreateDeploymentRequest) (domain.Deployment, error) {
	var appName string
	if err := r.pool.QueryRow(ctx, `SELECT name FROM applications WHERE id = $1`, req.ApplicationID).Scan(&appName); err != nil {
		return domain.Deployment{}, err
	}

	cluster := valueOr(req.Cluster, "k3s-prod-01")
	namespace := valueOr(req.Namespace, strings.TrimSuffix(appName, "-service"))
	version := valueOr(req.Version, "v0.1."+fmt.Sprint(int(randomByte()%9)+1))
	image := "harbor.local/sentinelops/" + appName

	item := domain.Deployment{
		ID:              "dep-" + randomHex(8),
		ApplicationID:   req.ApplicationID,
		ApplicationName: appName,
		Cluster:         cluster,
		Namespace:       namespace,
		Image:           image,
		Version:         version,
		SyncStatus:      "synced",
		HealthStatus:    "healthy",
	}

	if err := r.pool.QueryRow(ctx, `
		INSERT INTO deployments (id, application_id, cluster_name, namespace_name, image, version, sync_status, health_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING deployed_at`, item.ID, item.ApplicationID, item.Cluster, item.Namespace, item.Image, item.Version, item.SyncStatus, item.HealthStatus).Scan(&item.DeployedAt); err != nil {
		return domain.Deployment{}, err
	}

	_, _ = r.pool.Exec(ctx, `
		INSERT INTO observability_signals (id, source, signal_type, status, message)
		VALUES ($1, 'Argo CD', 'deployment', 'healthy', $2)`, "obs-"+randomHex(8), fmt.Sprintf("Deployment synced for %s on %s", appName, cluster))

	return item, nil
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

func (r *Repository) CreateSecurityAlert(ctx context.Context, req domain.CreateSecurityScanRequest) (domain.SecurityAlert, error) {
	source := valueOr(req.Source, "Trivy")
	severity := valueOr(req.Severity, "medium")
	title := valueOr(req.Title, defaultFindingTitle(source, severity))

	item := domain.SecurityAlert{
		ID:          "sec-" + randomHex(8),
		Source:      source,
		Severity:    severity,
		Title:       title,
		Application: req.Application,
		Status:      "open",
	}

	if err := r.pool.QueryRow(ctx, `
		INSERT INTO security_alerts (id, source, severity, title, application_name, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING detected_at`, item.ID, item.Source, item.Severity, item.Title, item.Application, item.Status).Scan(&item.DetectedAt); err != nil {
		return domain.SecurityAlert{}, err
	}

	_, _ = r.pool.Exec(ctx, `
		INSERT INTO observability_signals (id, source, signal_type, status, message)
		VALUES ($1, $2, 'security', 'warning', $3)`, "obs-"+randomHex(8), source, fmt.Sprintf("%s detected %s finding on %s", source, severity, item.Application))

	return item, nil
}

func (r *Repository) UpdateSecurityAlertStatus(ctx context.Context, id string, status string) (domain.SecurityAlert, error) {
	var item domain.SecurityAlert
	if err := r.pool.QueryRow(ctx, `
		UPDATE security_alerts
		SET status = $2
		WHERE id = $1
		RETURNING id, source, severity, title, application_name, status, detected_at`, id, status).Scan(&item.ID, &item.Source, &item.Severity, &item.Title, &item.Application, &item.Status, &item.DetectedAt); err != nil {
		return domain.SecurityAlert{}, err
	}
	return item, nil
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

func (r *Repository) ListRegistryArtifacts(ctx context.Context) ([]domain.RegistryArtifact, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, a.name, 'Harbor', d.image, d.version, 'generated', 'cosign-valid',
			CASE WHEN EXISTS (
				SELECT 1 FROM security_alerts s
				WHERE s.application_name = a.name
				AND s.status <> 'resolved'
				AND s.severity IN ('critical', 'high')
			) THEN 'failed' ELSE 'passed' END AS scan
		FROM deployments d
		JOIN applications a ON a.id = d.application_id
		ORDER BY d.deployed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.RegistryArtifact, 0)
	for rows.Next() {
		var item domain.RegistryArtifact
		if err := rows.Scan(&item.ID, &item.Name, &item.Registry, &item.Image, &item.Version, &item.SBOM, &item.Signature, &item.Scan); err != nil {
			return nil, err
		}
		item.Image = item.Image + ":" + item.Version
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CountApplications(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications`).Scan(&count)
	return count, err
}

func (r *Repository) CountClusters(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(COUNT(DISTINCT cluster_name), 0) FROM deployments`).Scan(&count)
	return count, err
}

func (r *Repository) CountDeployments(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deployments`).Scan(&count)
	return count, err
}

func (r *Repository) CountDeploymentsByHealth(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT health_status, COUNT(*) FROM deployments GROUP BY health_status`)
}

func (r *Repository) CountPipelinesByStatus(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT status, COUNT(*) FROM pipeline_runs GROUP BY status`)
}

func (r *Repository) CountSecurityBySeverity(ctx context.Context) (map[string]int, error) {
	return r.countBy(ctx, `SELECT severity, COUNT(*) FROM security_alerts WHERE status <> 'resolved' GROUP BY severity`)
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

func randomByte() byte {
	buf := make([]byte, 1)
	if _, err := rand.Read(buf); err != nil {
		return 42
	}
	return buf[0]
}

func valueOr(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func normalizeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func statusToSignal(status string) string {
	if status == "failed" {
		return "warning"
	}
	if status == "success" {
		return "healthy"
	}
	return "warning"
}

func defaultFindingTitle(source string, severity string) string {
	switch source {
	case "Trivy":
		return "Container image vulnerability detected"
	case "Semgrep":
		return "Static analysis rule matched risky code path"
	case "Gitleaks":
		return "Potential secret committed to repository"
	case "Falco":
		return "Unexpected runtime behavior detected"
	case "Wazuh":
		return "Host security event requires review"
	case "OPA-Gatekeeper":
		return "Kubernetes admission policy violation"
	default:
		return fmt.Sprintf("%s severity security finding", severity)
	}
}
