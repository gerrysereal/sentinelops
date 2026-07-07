# SentinelOps Security Model

## Security Objectives

SentinelOps is designed to act as a secure platform control plane. The platform must not become a bypass for GitOps, Kubernetes RBAC, or security policy controls.

## Local Development

Local development disables authentication with:

```env
AUTH_ENABLED=false
```

This makes the project runnable with Docker Compose. Production deployments must enable authentication.

## Production Authentication

Production deployments should use Keycloak as the OIDC identity provider.

Recommended controls:

- Validate JWT issuer against the configured Keycloak realm.
- Validate audience against the SentinelOps client ID.
- Use short-lived access tokens.
- Map Keycloak groups to SentinelOps roles.
- Deny all state-changing operations unless the user has a platform-admin or service-owner role.

## RBAC Model

Suggested roles:

- `platform-admin`: full administration.
- `security-engineer`: read security findings, update triage state.
- `sre`: read observability and deployments.
- `developer`: read own applications and request deployments.
- `viewer`: read-only access.

## Supply Chain Security

CI should include:

- Semgrep for SAST.
- Gitleaks for secret scanning.
- Trivy for dependency and container scanning.
- Syft for SBOM generation.
- Cosign for image signing.
- OPA Gatekeeper for deployment policy.

## Runtime Security

Recommended runtime controls:

- Falco for syscall-level runtime threat detection.
- Wazuh for SIEM and compliance signals.
- NetworkPolicy for service isolation.
- Pod Security Standards or equivalent admission control.
- Non-root containers.
- Read-only root filesystem where possible.
