package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "core_foundation_schema",
		SQL: `
		CREATE TABLE IF NOT EXISTS roles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS permissions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS role_permissions (
			role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			permission_id TEXT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (role_id, permission_id)
		);
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			role_id TEXT REFERENCES roles(id),
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS user_roles (
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			PRIMARY KEY (user_id, role_id)
		);
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			owner_name TEXT NOT NULL,
			environment TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS clusters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			provider TEXT NOT NULL,
			region TEXT NOT NULL DEFAULT '',
			kube_context TEXT NOT NULL DEFAULT '',
			api_server_url TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'unknown',
			health_message TEXT NOT NULL DEFAULT '',
			last_seen_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS applications (
			id TEXT PRIMARY KEY,
			project_id TEXT REFERENCES projects(id),
			name TEXT NOT NULL UNIQUE,
			owner_name TEXT NOT NULL,
			repository TEXT NOT NULL,
			environment TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS pipeline_runs (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			branch TEXT NOT NULL,
			commit_sha TEXT NOT NULL,
			status TEXT NOT NULL,
			stage TEXT NOT NULL,
			duration_seconds INT NOT NULL,
			finished_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			cluster_id TEXT REFERENCES clusters(id),
			cluster_name TEXT NOT NULL,
			namespace_name TEXT NOT NULL,
			image TEXT NOT NULL,
			version TEXT NOT NULL,
			sync_status TEXT NOT NULL,
			health_status TEXT NOT NULL,
			deployed_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS security_alerts (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			severity TEXT NOT NULL,
			title TEXT NOT NULL,
			application_name TEXT NOT NULL,
			status TEXT NOT NULL,
			detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS observability_signals (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			signal_type TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS notifications (
			id TEXT PRIMARY KEY,
			channel TEXT NOT NULL,
			target TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			actor TEXT NOT NULL DEFAULT 'system',
			action TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			message TEXT NOT NULL,
			request_id TEXT NOT NULL DEFAULT '',
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			environment TEXT NOT NULL DEFAULT 'global',
			is_sensitive BOOLEAN NOT NULL DEFAULT false,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS integrations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			tool_type TEXT NOT NULL,
			category TEXT NOT NULL,
			endpoint_url TEXT NOT NULL,
			access_token_encrypted TEXT NOT NULL DEFAULT '',
			username TEXT NOT NULL DEFAULT '',
			password_encrypted TEXT NOT NULL DEFAULT '',
			namespace TEXT NOT NULL DEFAULT '',
			tls_verify BOOLEAN NOT NULL DEFAULT true,
			sync_interval_seconds INT NOT NULL DEFAULT 60,
			enabled BOOLEAN NOT NULL DEFAULT false,
			status TEXT NOT NULL DEFAULT 'unknown',
			mode TEXT NOT NULL DEFAULT 'demo',
			last_health TEXT NOT NULL DEFAULT '',
			last_health_checked_at TIMESTAMPTZ,
			last_sync_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS integration_logs (
			id TEXT PRIMARY KEY,
			integration_id TEXT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
			action TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS secrets_metadata (
			id TEXT PRIMARY KEY,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			secret_name TEXT NOT NULL,
			provider TEXT NOT NULL DEFAULT 'postgres-encrypted',
			status TEXT NOT NULL DEFAULT 'active',
			rotated_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS connection_history (
			id TEXT PRIMARY KEY,
			integration_id TEXT NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
			action TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			latency_ms INT NOT NULL DEFAULT 0,
			checked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb
		);
		CREATE TABLE IF NOT EXISTS health_status (
			id TEXT PRIMARY KEY,
			component_type TEXT NOT NULL,
			component_id TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			checked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb
		);`,
	},
	{
		Version: 2,
		Name:    "indexes_and_compatibility_alters",
		SQL: `
		ALTER TABLE roles ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		ALTER TABLE users ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';
		ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		ALTER TABLE projects ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		ALTER TABLE clusters ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		ALTER TABLE settings ADD COLUMN IF NOT EXISTS environment TEXT NOT NULL DEFAULT 'global';
		ALTER TABLE settings ADD COLUMN IF NOT EXISTS is_sensitive BOOLEAN NOT NULL DEFAULT false;
		ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id TEXT NOT NULL DEFAULT '';
		ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;
		ALTER TABLE integrations ADD COLUMN IF NOT EXISTS last_health_checked_at TIMESTAMPTZ;
		ALTER TABLE clusters ADD COLUMN IF NOT EXISTS api_server_url TEXT NOT NULL DEFAULT '';
		ALTER TABLE clusters ADD COLUMN IF NOT EXISTS health_message TEXT NOT NULL DEFAULT '';
		ALTER TABLE clusters ADD COLUMN IF NOT EXISTS last_seen_at TIMESTAMPTZ;
		ALTER TABLE applications ADD COLUMN IF NOT EXISTS project_id TEXT REFERENCES projects(id);
		ALTER TABLE applications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
		ALTER TABLE deployments ADD COLUMN IF NOT EXISTS cluster_id TEXT REFERENCES clusters(id);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_integrations_name_lower ON integrations (lower(name));
		CREATE INDEX IF NOT EXISTS idx_integrations_status ON integrations(status);
		CREATE INDEX IF NOT EXISTS idx_integrations_enabled ON integrations(enabled);
		CREATE INDEX IF NOT EXISTS idx_integration_logs_integration_id ON integration_logs(integration_id);
		CREATE INDEX IF NOT EXISTS idx_connection_history_integration_id ON connection_history(integration_id, checked_at DESC);
		CREATE INDEX IF NOT EXISTS idx_health_status_component ON health_status(component_type, component_id, checked_at DESC);
		CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_pipeline_runs_application_id ON pipeline_runs(application_id);
		CREATE INDEX IF NOT EXISTS idx_deployments_application_id ON deployments(application_id);
		CREATE INDEX IF NOT EXISTS idx_security_alerts_detected_at ON security_alerts(detected_at DESC);
		CREATE INDEX IF NOT EXISTS idx_settings_environment ON settings(environment);`,
	},
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(ctx, pool, migration.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(ctx, pool, migration); err != nil {
			return err
		}
	}
	return nil
}

func migrationApplied(ctx context.Context, pool *pgxpool.Pool, version int) (bool, error) {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("check migration %d: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, migration Migration) error {
	start := time.Now()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", migration.Version, err)
	}
	defer tx.Rollback(ctx)

	for _, statement := range strings.Split(migration.SQL, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %d %s: %w", migration.Version, migration.Name, err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("record migration %d: %w", migration.Version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %d: %w", migration.Version, err)
	}
	slog.Info("database migration applied", "version", migration.Version, "name", migration.Name, "duration_ms", time.Since(start).Milliseconds())
	return nil
}
