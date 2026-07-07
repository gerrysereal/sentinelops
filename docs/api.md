# SentinelOps API Reference

Base URL for local development:

```text
http://localhost:8080/api/v1
```

## Health

```http
GET /health
```

Response:

```json
{
  "status": "ok",
  "service": "sentinelops-api"
}
```

## Overview

```http
GET /api/v1/overview
```

Returns dashboard summary metrics, status counts, recent alerts, resource usage, and integration health.

## Applications

```http
GET /api/v1/applications
```

Returns application inventory.

```http
POST /api/v1/applications
Content-Type: application/json

{
  "name": "orders-api",
  "owner": "commerce-team",
  "repository": "https://github.com/example/orders-api",
  "environment": "production"
}
```

## Pipelines

```http
GET /api/v1/pipelines
```

Returns CI/CD pipeline runs and scan stages.

## Deployments

```http
GET /api/v1/deployments
```

Returns GitOps deployment state.

## Security Alerts

```http
GET /api/v1/security/alerts
```

Returns security findings from scanner/runtime/security integrations.

## Observability Signals

```http
GET /api/v1/observability/signals
```

Returns metric, log, trace, and telemetry pipeline signals.

## Integrations

```http
GET /api/v1/integrations
```

Returns integration health for Argo CD, Harbor, Vault/OpenBao, Prometheus, Grafana, Loki, Tempo, Alloy, Falco, Wazuh, and Gatekeeper.

## Settings and Integrations API

SentinelOps treats the backend as an API gateway and integration hub. Integration credentials are encrypted before being persisted to PostgreSQL.

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v1/integrations` | List stored integration configurations without secret values. |
| POST | `/api/v1/integrations` | Add a new integration configuration. `accessToken` and `password` are encrypted server-side. |
| GET | `/api/v1/integrations/{id}` | Read one integration configuration. Secret values are never returned. |
| PUT | `/api/v1/integrations/{id}` | Edit an integration. Empty `accessToken` or `password` keeps the existing encrypted value. |
| DELETE | `/api/v1/integrations/{id}` | Delete an integration and cascade integration logs. |
| PATCH | `/api/v1/integrations/{id}/enabled` | Enable or disable an integration. |
| POST | `/api/v1/integrations/{id}/test` | Execute the selected adapter health check. |
| POST | `/api/v1/integrations/{id}/sync` | Execute adapter sync and write integration logs plus observability signal. |
| GET | `/api/v1/integrations/{id}/logs` | Return health, sync, update, and lifecycle logs for an integration. |

### Create Integration Example

```bash
curl -X POST http://localhost:8080/api/v1/integrations \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"Prometheus Local",
    "type":"Prometheus",
    "category":"Observability",
    "endpointUrl":"http://prometheus:9090",
    "namespace":"monitoring",
    "tlsVerify":true,
    "syncIntervalSeconds":60,
    "enabled":true
  }'
```

### Environment Modes

- `PLATFORM_MODE=demo`: uses real health-check adapters; useful for demos and offline classroom environments.
- `PLATFORM_MODE=lab`: uses HTTP adapters against Docker Compose or lab service endpoints.
- `PLATFORM_MODE=production`: uses HTTP adapters against real production endpoints.

