#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
API_BASE_URL=${API_BASE_URL:-http://localhost:8080}
GO_IMAGE=${GO_IMAGE:-golang:1.22-alpine}
VALIDATION_STARTED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)

log() {
  printf '%s\n' "[observability-validation] $*"
}

fail() {
  printf '%s\n' "[observability-validation] ERROR: $*" >&2
  exit 1
}

wait_for_api() {
  attempts=0
  until curl -fsS "$API_BASE_URL/health" >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ "$attempts" -ge 30 ]; then
      fail "API did not become healthy at $API_BASE_URL"
    fi
    sleep 1
  done
}

log "running backend tests"
docker run --rm \
  -v "$ROOT_DIR/apps/api:/src" \
  -w /src \
  "$GO_IMAGE" \
  sh -c 'go test ./... -count=1'

log "building and starting the API stack"
cd "$ROOT_DIR"
docker compose up -d --build postgres redis api
wait_for_api

log "checking health and readiness"
curl -fsS "$API_BASE_URL/health" >/dev/null
curl -fsS "$API_BASE_URL/ready" >/dev/null

log "generating HTTP, PostgreSQL, and Redis activity"
for _ in $(seq 1 10); do
  curl -fsS "$API_BASE_URL/health" >/dev/null
  curl -fsS "$API_BASE_URL/api/v1/applications" >/dev/null
  curl -fsS "$API_BASE_URL/api/v1/integrations" >/dev/null
  curl -fsS "$API_BASE_URL/api/v1/overview" >/dev/null
done

sleep 8

log "checking exporter and structured-log health"
recent_logs=$(docker compose logs --since="$VALIDATION_STARTED_AT" api 2>&1)
if printf '%s' "$recent_logs" | grep -Eqi 'traces export:.*(connection refused|Unavailable)|failed to upload metrics:.*(connection refused|Unavailable)'; then
  printf '%s\n' "$recent_logs" >&2
  fail "OpenTelemetry exporter reported a connectivity failure"
fi
if ! printf '%s' "$recent_logs" | grep -Eq 'trace_id[" ]*[:=][" ]*[0-9a-f]{32}'; then
  printf '%s\n' "$recent_logs" >&2
  fail "API request logs do not contain trace_id"
fi
if ! printf '%s' "$recent_logs" | grep -Eq 'span_id[" ]*[:=][" ]*[0-9a-f]{16}'; then
  printf '%s\n' "$recent_logs" >&2
  fail "API request logs do not contain span_id"
fi

log "automated validation passed"
printf '%s\n' ''
printf '%s\n' 'Manual Tempo checks:'
printf '%s\n' '  { resource.service.name = "sentinelops-api" }'
printf '%s\n' '  { resource.service.name = "sentinelops-api" && span.db.system = "postgresql" }'
printf '%s\n' '  { resource.service.name = "sentinelops-api" && span.db.system = "redis" }'
printf '%s\n' ''
printf '%s\n' 'Expected trace tree for /api/v1/overview:'
printf '%s\n' '  HTTP request'
printf '%s\n' '  ├── Redis GET/SET'
printf '%s\n' '  └── PostgreSQL SELECT spans on a cache miss'
