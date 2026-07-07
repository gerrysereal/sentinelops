# Deployment Guide

## Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

Services:

- `postgres`: platform data store.
- `redis`: dashboard cache.
- `api`: Go/Gin API.
- `web`: Next.js UI.

## Kubernetes

Apply the base manifests to a k3s cluster:

```bash
kubectl apply -f deploy/kubernetes/base/namespace.yaml
kubectl apply -f deploy/kubernetes/base/
```

The base manifests are intentionally simple and are suitable for local k3s, k3d, or minikube-style clusters.

## Helm

```bash
helm upgrade --install sentinelops deploy/helm/sentinelops -n sentinelops --create-namespace
```

## GitOps with Argo CD

```bash
kubectl apply -f gitops/argocd/sentinelops-application.yaml
```

Edit the repository URL in the Argo CD application manifest before using it in a real cluster.
