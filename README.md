# SentinelOps

SentinelOps is an open-source Internal Developer Platform (IDP) that unifies DevOps, DevSecOps, GitOps, observability, runtime security, and platform operations into one web application.

This repository is intentionally built as a **runnable production-style MVP**. It uses real service boundaries, modular code, Docker support, Kubernetes manifests, Helm chart scaffolding, GitHub Actions, and documentation. External DevSecOps tools such as Argo CD, Harbor, Vault/OpenBao, Trivy, Semgrep, Gitleaks, Falco, Wazuh, OPA Gatekeeper, Cosign, Syft, Prometheus, Grafana, Loki, Tempo, and Grafana Alloy are represented as integration modules and deployment hooks so the project can run locally without requiring a full enterprise cluster on day one.

## Tech Stack

### Frontend
- Next.js
- React
- TypeScript
- TailwindCSS
- shadcn/ui-style components
- TanStack Query
- Zustand

### Backend
- Go
- Gin
- PostgreSQL
- Redis

### Platform and DevSecOps
- Docker / Docker Compose
- Kubernetes / k3s-compatible manifests
- Helm
- Argo CD application manifest
- GitHub Actions CI
- Security scanner workflow placeholders for Trivy, Semgrep, and Gitleaks

## Quick Start

### Prerequisites

- Docker
- Docker Compose

### Run all services

```bash
cp .env.example .env
docker compose up --build
```

Open:

- Web UI: http://localhost:3000
- API health: http://localhost:8080/health
- API overview: http://localhost:8080/api/v1/overview

### Default local behavior

Authentication is disabled by default for local development through `AUTH_ENABLED=false`. The backend still includes middleware boundaries so Keycloak/OIDC can be enabled later without changing handlers.

## Local Development

### Backend

```bash
cd apps/api
go mod download
go run ./cmd/api
```

### Frontend

```bash
cd apps/web
npm install
npm run dev
```

## Repository Layout

```text
sentinelops/
  apps/
    api/                  Go API service
    web/                  Next.js frontend
  deploy/
    docker/               Dockerfiles
    kubernetes/base/      k3s-compatible manifests
    helm/sentinelops/     Helm chart scaffold
  docs/                   Architecture and operating documentation
  gitops/argocd/          Argo CD application manifest
  iac/opentofu/           OpenTofu starter module
  ansible/                Ansible starter playbook
  .github/workflows/      CI and security workflows
```

## Core Features Implemented

- Platform dashboard with deployment, security, registry, pipeline, and observability summaries
- Application inventory
- CI/CD pipeline status
- GitOps deployment status
- Security alert feed
- Observability signal feed
- Integration health page
- PostgreSQL-backed seed data
- Redis-backed overview cache
- Docker Compose runtime
- Kubernetes manifests and Helm scaffold
- CI workflow with lint/test/security stages

## Architecture Documentation

See:

- [Architecture](docs/architecture.md)
- [Security Model](docs/security.md)
- [Deployment Guide](docs/deployment.md)
- [API Reference](docs/api.md)
- [Operational Runbook](docs/runbook.md)

## Design Philosophy

SentinelOps follows a modular platform architecture. The first production milestone focuses on a reliable IDP control plane and integration model before attempting to run every external tool inside the same repository. This avoids unnecessary local complexity while preserving a clear path toward enterprise deployment.

## Production-Oriented Settings Module

The Settings page is backed by the API and PostgreSQL. It supports integration lifecycle operations:

- Add Integration
- Edit Integration
- Delete Integration
- Enable / Disable Integration
- Test Connection
- Sync Data
- View Health
- View Last Sync
- View Integration Logs

Credentials are encrypted before storage. Set `INTEGRATION_ENCRYPTION_KEY` in `.env` before running anything beyond local demo mode.

```bash
PLATFORM_MODE=demo docker-compose up --build
```

Switching modes:

```env
PLATFORM_MODE=demo        # real health-check adapters
PLATFORM_MODE=lab         # HTTP adapters for Docker Compose/lab services
PLATFORM_MODE=production  # HTTP adapters for real endpoints
```

