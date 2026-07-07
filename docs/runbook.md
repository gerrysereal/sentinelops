# SentinelOps Operational Runbook

## API does not start

1. Check PostgreSQL health.
2. Check `DATABASE_URL`.
3. Check container logs:

```bash
docker compose logs api
```

## Web cannot load data

1. Confirm API health:

```bash
curl http://localhost:8080/health
```

2. Confirm browser can access API:

```bash
curl http://localhost:8080/api/v1/overview
```

3. Check `NEXT_PUBLIC_API_BASE_URL`.

## Reset local data

```bash
docker compose down -v
docker compose up --build
```

## Kubernetes rollout check

```bash
kubectl -n sentinelops get pods
kubectl -n sentinelops logs deploy/sentinelops-api
kubectl -n sentinelops logs deploy/sentinelops-web
```
