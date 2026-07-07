package database

import (
	"context"
	"crypto/rand"
	stdsql "database/sql"
	"encoding/hex"
	"encoding/json"
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
	image := appName

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

func (r *Repository) HasEnabledIntegrationType(ctx context.Context, toolTypes ...string) (bool, error) {
	if len(toolTypes) == 0 {
		return false, nil
	}
	placeholders := make([]string, 0, len(toolTypes))
	args := make([]any, 0, len(toolTypes))
	for index, toolType := range toolTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, toolType)
	}
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM integrations WHERE enabled = true AND tool_type IN (%s))`, strings.Join(placeholders, ","))
	var exists bool
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) CountApplications(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications`).Scan(&count)
	return count, err
}

func (r *Repository) CountClusters(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM clusters`).Scan(&count)
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

func (r *Repository) ListIntegrations(ctx context.Context) ([]domain.IntegrationConfig, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''
		FROM integrations
		ORDER BY category, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.IntegrationConfig, 0)
	for rows.Next() {
		item, err := scanIntegrationPublic(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetIntegration(ctx context.Context, id string) (domain.IntegrationConfig, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> '', access_token_encrypted, password_encrypted
		FROM integrations
		WHERE id = $1`, id)
	return scanIntegrationPrivate(row)
}

func (r *Repository) CreateIntegration(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationConfig, error) {
	if cfg.ID == "" {
		cfg.ID = "int-" + randomHex(8)
	}
	if cfg.Status == "" {
		if cfg.Enabled {
			cfg.Status = domain.IntegrationStatusUnknown
		} else {
			cfg.Status = domain.IntegrationStatusDisabled
		}
	}
	if cfg.SyncIntervalSeconds == 0 {
		cfg.SyncIntervalSeconds = 60
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO integrations
		(id, name, tool_type, category, endpoint_url, access_token_encrypted, username, password_encrypted, namespace,
		 tls_verify, sync_interval_seconds, enabled, status, mode, last_health)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''`,
		cfg.ID, cfg.Name, cfg.Type, cfg.Category, cfg.EndpointURL, cfg.AccessToken, cfg.Username, cfg.Password, cfg.Namespace,
		cfg.TLSVerify, cfg.SyncIntervalSeconds, cfg.Enabled, cfg.Status, cfg.Mode, cfg.Health)

	created, err := scanIntegrationPublic(row)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	_ = r.CreateIntegrationLog(ctx, created.ID, "create", created.Status, "integration configuration saved")
	_ = r.CreateAuditLog(ctx, "system", "create", "integration", created.ID, "integration created: "+created.Name)
	if cfg.AccessToken != "" {
		_ = r.UpsertSecretMetadata(ctx, "integration", created.ID, "access_token")
	}
	if cfg.Password != "" {
		_ = r.UpsertSecretMetadata(ctx, "integration", created.ID, "password")
	}
	return created, nil
}

func (r *Repository) UpdateIntegration(ctx context.Context, cfg domain.IntegrationConfig) (domain.IntegrationConfig, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE integrations
		SET name = $2,
			category = $3,
			endpoint_url = $4,
			access_token_encrypted = $5,
			username = $6,
			password_encrypted = $7,
			namespace = $8,
			tls_verify = $9,
			sync_interval_seconds = $10,
			enabled = $11,
			status = $12,
			mode = $13,
			last_health = $14,
			last_sync_at = $15,
			updated_at = now()
		WHERE id = $1
		RETURNING id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''`,
		cfg.ID, cfg.Name, cfg.Category, cfg.EndpointURL, cfg.AccessToken, cfg.Username, cfg.Password, cfg.Namespace,
		cfg.TLSVerify, cfg.SyncIntervalSeconds, cfg.Enabled, cfg.Status, cfg.Mode, cfg.Health, cfg.LastSyncAt)
	updated, err := scanIntegrationPublic(row)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	_ = r.CreateIntegrationLog(ctx, updated.ID, "update", updated.Status, "integration configuration updated")
	_ = r.CreateAuditLog(ctx, "system", "update", "integration", updated.ID, "integration updated: "+updated.Name)
	if cfg.AccessToken != "" {
		_ = r.UpsertSecretMetadata(ctx, "integration", updated.ID, "access_token")
	}
	if cfg.Password != "" {
		_ = r.UpsertSecretMetadata(ctx, "integration", updated.ID, "password")
	}
	return updated, nil
}

func (r *Repository) DeleteIntegration(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM integrations WHERE id = $1`, id)
	if err == nil {
		_ = r.CreateAuditLog(ctx, "system", "delete", "integration", id, "integration deleted")
	}
	return err
}

func (r *Repository) SetIntegrationEnabled(ctx context.Context, id string, enabled bool) (domain.IntegrationConfig, error) {
	status := domain.IntegrationStatusDisabled
	message := "integration disabled"
	if enabled {
		status = domain.IntegrationStatusUnknown
		message = "integration enabled"
	}
	row := r.pool.QueryRow(ctx, `
		UPDATE integrations
		SET enabled = $2, status = $3, last_health = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''`, id, enabled, status, message)
	updated, err := scanIntegrationPublic(row)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	_ = r.CreateIntegrationLog(ctx, id, "set_enabled", status, message)
	_ = r.CreateAuditLog(ctx, "system", "set_enabled", "integration", id, message)
	return updated, nil
}

func (r *Repository) UpdateIntegrationHealth(ctx context.Context, id string, health domain.IntegrationHealth) (domain.IntegrationConfig, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE integrations
		SET status = $2, last_health = $3, last_health_checked_at = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''`, id, health.Status, health.Message, health.CheckedAt)
	updated, err := scanIntegrationPublic(row)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	_ = r.CreateIntegrationLog(ctx, id, "health", health.Status, health.Message)
	_ = r.CreateConnectionHistory(ctx, id, "health", health.Status, health.Message, health.LatencyMs, health.Attributes)
	_ = r.CreateHealthStatus(ctx, "integration", id, health.Status, health.Message, health.Attributes)
	return updated, nil
}

func (r *Repository) MarkIntegrationSynced(ctx context.Context, id string, sync domain.IntegrationSyncResult) (domain.IntegrationConfig, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE integrations
		SET status = $2, last_health = $3, last_sync_at = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, name, tool_type, category, endpoint_url, username, namespace, tls_verify,
			sync_interval_seconds, enabled, status, mode, last_health, last_sync_at, created_at, updated_at,
			access_token_encrypted <> '', password_encrypted <> ''`, id, sync.Status, sync.Message, sync.SyncedAt)
	updated, err := scanIntegrationPublic(row)
	if err != nil {
		return domain.IntegrationConfig{}, err
	}
	_ = r.CreateIntegrationLog(ctx, id, "sync", sync.Status, sync.Message)
	metadata := map[string]string{}
	for key, value := range sync.Metadata {
		metadata[key] = value
	}
	for key, value := range sync.Resources {
		metadata["resource_"+key] = fmt.Sprintf("%d", value)
	}
	_ = r.CreateConnectionHistory(ctx, id, "sync", sync.Status, sync.Message, 0, metadata)
	return updated, nil
}

func (r *Repository) CreateIntegrationLog(ctx context.Context, integrationID string, action string, status string, message string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO integration_logs (id, integration_id, action, status, message)
		VALUES ($1, $2, $3, $4, $5)`, "ilog-"+randomHex(8), integrationID, action, status, message)
	return err
}

func (r *Repository) ListIntegrationLogs(ctx context.Context, integrationID string) ([]domain.IntegrationLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, integration_id, action, status, message, created_at
		FROM integration_logs
		WHERE integration_id = $1
		ORDER BY created_at DESC
		LIMIT 100`, integrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.IntegrationLog, 0)
	for rows.Next() {
		var item domain.IntegrationLog
		if err := rows.Scan(&item.ID, &item.IntegrationID, &item.Action, &item.Status, &item.Message, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateAuditLog(ctx context.Context, actor string, action string, resourceType string, resourceID string, message string) error {
	if actor == "" || actor == "system" {
		actor = contextValue(ctx, "actor", actor)
	}
	requestID := contextValue(ctx, "request_id", "")
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (id, actor, action, resource_type, resource_id, message, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, "audit-"+randomHex(8), actor, action, resourceType, resourceID, message, requestID)
	return err
}

func (r *Repository) CreateObservabilitySignal(ctx context.Context, source string, signalType string, status string, message string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO observability_signals (id, source, signal_type, status, message)
		VALUES ($1, $2, $3, $4, $5)`, "obs-"+randomHex(8), source, signalType, status, message)
	return err
}

type integrationPublicScanner interface {
	Scan(dest ...any) error
}

func scanIntegrationPublic(row integrationPublicScanner) (domain.IntegrationConfig, error) {
	var item domain.IntegrationConfig
	var lastSync stdsql.NullTime
	if err := row.Scan(&item.ID, &item.Name, &item.Type, &item.Category, &item.EndpointURL, &item.Username, &item.Namespace,
		&item.TLSVerify, &item.SyncIntervalSeconds, &item.Enabled, &item.Status, &item.Mode, &item.Health, &lastSync,
		&item.CreatedAt, &item.UpdatedAt, &item.HasAccessToken, &item.HasPassword); err != nil {
		return domain.IntegrationConfig{}, err
	}
	if lastSync.Valid {
		item.LastSyncAt = &lastSync.Time
	}
	return item, nil
}

func scanIntegrationPrivate(row integrationPublicScanner) (domain.IntegrationConfig, error) {
	var item domain.IntegrationConfig
	var lastSync stdsql.NullTime
	if err := row.Scan(&item.ID, &item.Name, &item.Type, &item.Category, &item.EndpointURL, &item.Username, &item.Namespace,
		&item.TLSVerify, &item.SyncIntervalSeconds, &item.Enabled, &item.Status, &item.Mode, &item.Health, &lastSync,
		&item.CreatedAt, &item.UpdatedAt, &item.HasAccessToken, &item.HasPassword, &item.AccessToken, &item.Password); err != nil {
		return domain.IntegrationConfig{}, err
	}
	if lastSync.Valid {
		item.LastSyncAt = &lastSync.Time
	}
	return item, nil
}

func (r *Repository) ListSettings(ctx context.Context) ([]domain.Setting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT key, CASE WHEN is_sensitive THEN '' ELSE value END, environment, is_sensitive, updated_at
		FROM settings
		ORDER BY environment, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Setting, 0)
	for rows.Next() {
		var item domain.Setting
		if err := rows.Scan(&item.Key, &item.Value, &item.Environment, &item.IsSensitive, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertSetting(ctx context.Context, key string, req domain.UpdateSettingRequest) (domain.Setting, error) {
	environment := valueOr(req.Environment, "global")
	row := r.pool.QueryRow(ctx, `
		INSERT INTO settings (key, value, environment, is_sensitive, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, environment = EXCLUDED.environment, is_sensitive = EXCLUDED.is_sensitive, updated_at = now()
		RETURNING key, CASE WHEN is_sensitive THEN '' ELSE value END, environment, is_sensitive, updated_at`, key, req.Value, environment, req.IsSensitive)
	var item domain.Setting
	if err := row.Scan(&item.Key, &item.Value, &item.Environment, &item.IsSensitive, &item.UpdatedAt); err != nil {
		return domain.Setting{}, err
	}
	_ = r.CreateAuditLog(ctx, "system", "upsert", "setting", key, "setting updated")
	return item, nil
}

func (r *Repository) GetIntegrationStatus(ctx context.Context, id string) (domain.IntegrationStatusView, error) {
	var item domain.IntegrationStatusView
	var lastSync stdsql.NullTime
	var lastChecked stdsql.NullTime
	if err := r.pool.QueryRow(ctx, `
		SELECT id, status, last_health, last_sync_at, last_health_checked_at, enabled
		FROM integrations
		WHERE id = $1`, id).Scan(&item.ID, &item.Status, &item.Health, &lastSync, &lastChecked, &item.Enabled); err != nil {
		return domain.IntegrationStatusView{}, err
	}
	if lastSync.Valid {
		item.LastSyncAt = &lastSync.Time
	}
	if lastChecked.Valid {
		item.LastCheckedAt = &lastChecked.Time
	}
	return item, nil
}

func (r *Repository) CreateConnectionHistory(ctx context.Context, integrationID string, action string, status string, message string, latencyMs int, metadata map[string]string) error {
	payload, _ := json.Marshal(metadata)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO connection_history (id, integration_id, action, status, message, latency_ms, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)`, "conn-"+randomHex(8), integrationID, action, status, message, latencyMs, string(payload))
	return err
}

func (r *Repository) ListConnectionHistory(ctx context.Context, integrationID string) ([]domain.ConnectionHistory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, integration_id, action, status, message, latency_ms, checked_at
		FROM connection_history
		WHERE integration_id = $1
		ORDER BY checked_at DESC
		LIMIT 100`, integrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ConnectionHistory, 0)
	for rows.Next() {
		var item domain.ConnectionHistory
		if err := rows.Scan(&item.ID, &item.IntegrationID, &item.Action, &item.Status, &item.Message, &item.LatencyMs, &item.CheckedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateHealthStatus(ctx context.Context, componentType string, componentID string, status string, message string, metadata map[string]string) error {
	payload, _ := json.Marshal(metadata)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO health_status (id, component_type, component_id, status, message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)`, "health-"+randomHex(8), componentType, componentID, status, message, string(payload))
	return err
}

func (r *Repository) UpsertSecretMetadata(ctx context.Context, resourceType string, resourceID string, secretName string) error {
	id := "secret-" + normalizeName(resourceType) + "-" + normalizeName(resourceID) + "-" + normalizeName(secretName)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO secrets_metadata (id, resource_type, resource_id, secret_name, provider, status, rotated_at, updated_at)
		VALUES ($1, $2, $3, $4, 'postgres-encrypted', 'active', now(), now())
		ON CONFLICT (id) DO UPDATE SET status = 'active', rotated_at = now(), updated_at = now()`, id, resourceType, resourceID, secretName)
	return err
}

func contextValue(ctx context.Context, key string, fallback string) string {
	if value, ok := ctx.Value(key).(string); ok && value != "" {
		return value
	}
	return fallback
}
