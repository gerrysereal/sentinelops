package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	if err := seedCorePlatformData(ctx, pool); err != nil {
		return err
	}
	if err := seedIntegrationCatalog(ctx, pool); err != nil {
		return err
	}
	return seedDefaultSettings(ctx, pool)
}

func seedCorePlatformData(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM applications`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	batch := []string{
		`INSERT INTO applications (id, name, owner_name, repository, environment, status) VALUES
		('app-checkout', 'checkout-service', 'platform-team', 'https://github.com/example/checkout-service', 'production', 'healthy'),
		('app-payments', 'payments-api', 'payments-team', 'https://github.com/example/payments-api', 'production', 'degraded'),
		('app-inventory', 'inventory-worker', 'supply-team', 'https://github.com/example/inventory-worker', 'staging', 'healthy'),
		('app-portal', 'developer-portal', 'platform-team', 'https://github.com/example/developer-portal', 'production', 'healthy');`,
		`INSERT INTO pipeline_runs (id, application_id, branch, commit_sha, status, stage, duration_seconds, finished_at) VALUES
		('pipe-001', 'app-checkout', 'main', 'a9f31c2', 'success', 'push-image', 412, now() - interval '15 minutes'),
		('pipe-002', 'app-payments', 'main', 'b71da3e', 'failed', 'semgrep-sast', 187, now() - interval '31 minutes'),
		('pipe-003', 'app-inventory', 'release/1.8', 'e83ac91', 'running', 'trivy-image-scan', 256, now() - interval '6 minutes'),
		('pipe-004', 'app-portal', 'main', 'c11be20', 'success', 'argocd-sync', 519, now() - interval '52 minutes'),
		('pipe-005', 'app-checkout', 'hotfix/cart', 'fae183a', 'pending', 'approval', 0, now() - interval '2 minutes');`,
		`INSERT INTO deployments (id, application_id, cluster_name, namespace_name, image, version, sync_status, health_status, deployed_at) VALUES
		('dep-001', 'app-checkout', 'k3s-prod-01', 'checkout', 'harbor.local/sentinelops/checkout-service', '1.4.2', 'synced', 'healthy', now() - interval '1 hour'),
		('dep-002', 'app-payments', 'k3s-prod-01', 'payments', 'harbor.local/sentinelops/payments-api', '2.1.0', 'out-of-sync', 'degraded', now() - interval '3 hours'),
		('dep-003', 'app-inventory', 'k3s-staging-01', 'inventory', 'harbor.local/sentinelops/inventory-worker', '1.8.0-rc2', 'synced', 'progressing', now() - interval '25 minutes'),
		('dep-004', 'app-portal', 'k3s-prod-01', 'platform', 'harbor.local/sentinelops/developer-portal', '0.9.5', 'synced', 'healthy', now() - interval '4 hours');`,
		`INSERT INTO security_alerts (id, source, severity, title, application_name, status, detected_at) VALUES
		('sec-001', 'Trivy', 'high', 'Base image contains critical OpenSSL CVE', 'payments-api', 'open', now() - interval '10 minutes'),
		('sec-002', 'Gitleaks', 'medium', 'Potential API token pattern detected', 'checkout-service', 'triaged', now() - interval '27 minutes'),
		('sec-003', 'Falco', 'high', 'Unexpected shell spawned in container', 'payments-api', 'open', now() - interval '4 minutes'),
		('sec-004', 'OPA Gatekeeper', 'low', 'Missing required cost-center label', 'inventory-worker', 'resolved', now() - interval '2 hours'),
		('sec-005', 'Semgrep', 'medium', 'Unsafe SQL construction pattern', 'developer-portal', 'open', now() - interval '48 minutes');`,
		`INSERT INTO observability_signals (id, source, signal_type, status, message, created_at) VALUES
		('obs-001', 'Prometheus', 'metric', 'healthy', 'Cluster CPU usage is within expected threshold', now() - interval '5 minutes'),
		('obs-002', 'Loki', 'log', 'warning', 'payments-api error rate increased in the last 15 minutes', now() - interval '9 minutes'),
		('obs-003', 'Tempo', 'trace', 'warning', 'checkout-service p95 latency above SLO target', now() - interval '14 minutes'),
		('obs-004', 'Grafana Alloy', 'agent', 'healthy', 'Telemetry pipeline is active for 4 namespaces', now() - interval '3 minutes');`,
	}

	for _, statement := range batch {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func seedIntegrationCatalog(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM integrations`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, err := pool.Exec(ctx, `INSERT INTO integrations
		(id, name, tool_type, category, endpoint_url, namespace, tls_verify, sync_interval_seconds, enabled, status, mode, last_health)
		VALUES
		('int-prometheus', 'Prometheus', 'Prometheus', 'Observability', 'http://prometheus:9090', 'monitoring', true, 60, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-grafana', 'Grafana', 'Grafana', 'Observability', 'http://grafana:3000', 'monitoring', true, 60, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-harbor', 'Harbor', 'Harbor', 'Registry', 'http://harbor.local', 'registry', true, 300, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-argocd', 'Argo CD', 'ArgoCD', 'GitOps', 'http://argocd-server.argocd.svc.cluster.local', 'argocd', true, 60, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-vault', 'Vault/OpenBao', 'Vault', 'Secrets', 'http://vault:8200', 'vault', true, 300, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-github', 'GitHub', 'GitHub', 'SCM', 'https://api.github.com', '', true, 300, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-sonarqube', 'SonarQube', 'SonarQube', 'Code Quality', 'http://sonarqube:9000', 'quality', true, 300, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-wazuh', 'Wazuh', 'Wazuh', 'SIEM', 'http://wazuh-manager:55000', 'security', true, 300, false, 'disabled', 'demo', 'Not enabled yet'),
		('int-falco', 'Falco', 'Falco', 'Runtime Security', 'http://falco-exporter:9376', 'security', true, 60, false, 'disabled', 'demo', 'Not enabled yet')
		ON CONFLICT (id) DO NOTHING;`)
	return err
}

func seedDefaultSettings(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `INSERT INTO settings (key, value) VALUES
		('environment_mode', 'demo'),
		('default_namespace', 'sentinelops'),
		('audit_retention_days', '90')
		ON CONFLICT (key) DO NOTHING;`)
	return err
}
