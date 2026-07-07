package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Seed(ctx context.Context, pool *pgxpool.Pool, platformMode string) error {
	if err := seedRBAC(ctx, pool); err != nil {
		return err
	}
	return seedDefaultSettings(ctx, pool, platformMode)
}

func seedRBAC(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`INSERT INTO roles (id, name, description) VALUES
		('role-platform-admin', 'platform-admin', 'Full SentinelOps administration'),
		('role-platform-engineer', 'platform-engineer', 'Platform operations and integration management'),
		('role-security-engineer', 'security-engineer', 'Security findings and policy operations'),
		('role-developer', 'developer', 'Application delivery workflows'),
		('role-viewer', 'viewer', 'Read-only access')
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, updated_at = now();`,
		`INSERT INTO permissions (id, name, description) VALUES
		('perm-platform-read', 'platform:read', 'Read platform overview and health'),
		('perm-settings-read', 'settings:read', 'Read platform settings'),
		('perm-settings-write', 'settings:write', 'Update platform settings'),
		('perm-integration-read', 'integration:read', 'Read integration configurations'),
		('perm-integration-write', 'integration:write', 'Create and update integration configurations'),
		('perm-integration-delete', 'integration:delete', 'Delete integration configurations'),
		('perm-integration-operate', 'integration:operate', 'Test, sync, enable, and disable integrations'),
		('perm-application-read', 'application:read', 'Read applications'),
		('perm-application-write', 'application:write', 'Create and update applications'),
		('perm-pipeline-operate', 'pipeline:operate', 'Operate pipeline actions'),
		('perm-deployment-operate', 'deployment:operate', 'Operate deployment actions'),
		('perm-security-operate', 'security:operate', 'Operate security actions'),
		('perm-audit-read', 'audit:read', 'Read audit logs')
		ON CONFLICT (id) DO NOTHING;`,
		`INSERT INTO role_permissions (role_id, permission_id)
		SELECT 'role-platform-admin', id FROM permissions
		ON CONFLICT DO NOTHING;`,
		`INSERT INTO role_permissions (role_id, permission_id) VALUES
		('role-platform-engineer', 'perm-platform-read'),
		('role-platform-engineer', 'perm-settings-read'),
		('role-platform-engineer', 'perm-integration-read'),
		('role-platform-engineer', 'perm-integration-write'),
		('role-platform-engineer', 'perm-integration-operate'),
		('role-platform-engineer', 'perm-application-read'),
		('role-platform-engineer', 'perm-application-write'),
		('role-platform-engineer', 'perm-pipeline-operate'),
		('role-platform-engineer', 'perm-deployment-operate'),
		('role-security-engineer', 'perm-platform-read'),
		('role-security-engineer', 'perm-integration-read'),
		('role-security-engineer', 'perm-security-operate'),
		('role-developer', 'perm-platform-read'),
		('role-developer', 'perm-application-read'),
		('role-developer', 'perm-pipeline-operate'),
		('role-viewer', 'perm-platform-read'),
		('role-viewer', 'perm-settings-read'),
		('role-viewer', 'perm-integration-read'),
		('role-viewer', 'perm-application-read')
		ON CONFLICT DO NOTHING;`,
		`INSERT INTO users (id, email, display_name, role_id, status) VALUES
		('user-local-admin', 'local-admin@sentinelops.internal', 'Local Platform Admin', 'role-platform-admin', 'active')
		ON CONFLICT (id) DO UPDATE SET role_id = EXCLUDED.role_id, status = EXCLUDED.status, updated_at = now();`,
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func seedDefaultSettings(ctx context.Context, pool *pgxpool.Pool, platformMode string) error {
	_, err := pool.Exec(ctx, `INSERT INTO settings (key, value, environment, is_sensitive) VALUES
		('environment_mode', $1, 'global', false),
		('default_namespace', 'sentinelops', 'global', false),
		('audit_retention_days', '90', 'global', false),
		('integration_retry_policy', 'exponential-backoff', 'global', false)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, environment = EXCLUDED.environment, updated_at = now();`, platformMode)
	return err
}
