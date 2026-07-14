# SentinelOps MVP Observability

SentinelOps MVP owns application instrumentation. SentinelOps Lab owns the
OpenTelemetry Collector, Tempo, Prometheus, Grafana, Loki, Kubernetes, and
GitOps deployment resources.

## Implemented instrumentation

The API exports OTLP telemetry with `service.name=sentinelops-api`.

- Gin HTTP server spans for every route.
- PostgreSQL client spans and pool metrics through `otelpgx`.
- Redis client spans and metrics through `redisotel`.
- Structured HTTP request logs correlated with `request_id`, `actor`,
  `trace_id`, and `span_id`.
- Graceful trace and metric flushing during API shutdown.

Sensitive database statements, Redis command arguments, credentials, tokens,
and connection strings are not exported.

## Runtime configuration

```env
OTEL_SDK_DISABLED=false
OTEL_SERVICE_NAME=sentinelops-api
OTEL_SERVICE_VERSION=0.1.0
OTEL_DEPLOYMENT_ENVIRONMENT=lab
OTEL_EXPORTER_OTLP_ENDPOINT=http://172.17.0.1:4317
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_EXPORTER_OTLP_TIMEOUT=10000
OTEL_METRIC_EXPORT_INTERVAL=30000
OTEL_BSP_SCHEDULE_DELAY=5000
OTEL_TRACES_SAMPLER=parentbased_always_on
OTEL_TRACES_SAMPLER_ARG=1
```

OpenTelemetry duration environment variables use milliseconds when supplied as
integers.

## Docker-to-k3s development bridge

A Docker-hosted API needs an address that reaches the Collector in k3s. For
local validation, keep this process running:

```bash
kubectl -n monitoring port-forward \
  --address 172.17.0.1 \
  svc/otel-opentelemetry-collector \
  4317:4317
```

The Collector OTLP receiver must bind to all pod interfaces:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
```

Port-forward is for local debugging only. GitOps deployments should inject the
in-cluster Collector service endpoint directly into the SentinelOps API pod.

## Automated validation

Run from the repository root:

```bash
make validate-observability
```

The validation performs backend tests, builds and starts the API, checks health
and readiness, generates HTTP/PostgreSQL/Redis activity, rejects recent OTLP
connectivity failures, and verifies trace/span correlation in API logs.

## Tempo validation

In Grafana Explore, select Tempo and use TraceQL.

All SentinelOps API traces:

```traceql
{ resource.service.name = "sentinelops-api" }
```

PostgreSQL spans:

```traceql
{ resource.service.name = "sentinelops-api" && span.db.system = "postgresql" }
```

Redis spans:

```traceql
{ resource.service.name = "sentinelops-api" && span.db.system = "redis" }
```

Generate cache activity with:

```bash
for i in $(seq 1 10); do
  curl -s http://localhost:8080/api/v1/overview >/dev/null
done
```

A cache miss should show Redis and PostgreSQL child spans. A cache hit should
show Redis without repeating the database aggregation queries.

## Log correlation

Request logs include fields similar to:

```text
request_id=... actor=local-admin trace_id=... span_id=... method=GET path=/api/v1/overview status=200
```

Use the `trace_id` value to open the corresponding trace in Tempo.

## Reboot recovery

Docker services use `restart: unless-stopped`, but `kubectl port-forward` is a
temporary process and must be restarted after a reboot. A missing port-forward
produces OTLP connection-refused messages while the API itself remains healthy.
