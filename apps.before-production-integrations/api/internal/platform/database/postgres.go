package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS applications (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			owner_name TEXT NOT NULL,
			repository TEXT NOT NULL,
			environment TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS pipeline_runs (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			branch TEXT NOT NULL,
			commit_sha TEXT NOT NULL,
			status TEXT NOT NULL,
			stage TEXT NOT NULL,
			duration_seconds INT NOT NULL,
			finished_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			cluster_name TEXT NOT NULL,
			namespace_name TEXT NOT NULL,
			image TEXT NOT NULL,
			version TEXT NOT NULL,
			sync_status TEXT NOT NULL,
			health_status TEXT NOT NULL,
			deployed_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS security_alerts (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			severity TEXT NOT NULL,
			title TEXT NOT NULL,
			application_name TEXT NOT NULL,
			status TEXT NOT NULL,
			detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE TABLE IF NOT EXISTS observability_signals (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			signal_type TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_pipeline_runs_application_id ON pipeline_runs(application_id);`,
		`CREATE INDEX IF NOT EXISTS idx_deployments_application_id ON deployments(application_id);`,
		`CREATE INDEX IF NOT EXISTS idx_security_alerts_detected_at ON security_alerts(detected_at DESC);`,
	}

	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
