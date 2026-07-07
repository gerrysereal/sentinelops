package domain

import "time"

const (
	IntegrationStatusUnknown      = "unknown"
	IntegrationStatusConnected    = "connected"
	IntegrationStatusDisconnected = "disconnected"
	IntegrationStatusDisabled     = "disabled"
	IntegrationStatusSyncing      = "syncing"
	IntegrationStatusError        = "error"
	IntegrationModeDemo           = "demo"
	IntegrationModeLab            = "lab"
	IntegrationModeProduction     = "production"
)

type IntegrationConfig struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Type                string     `json:"type"`
	Category            string     `json:"category"`
	EndpointURL         string     `json:"endpointUrl"`
	Username            string     `json:"username,omitempty"`
	Namespace           string     `json:"namespace,omitempty"`
	TLSVerify           bool       `json:"tlsVerify"`
	SyncIntervalSeconds int        `json:"syncIntervalSeconds"`
	Enabled             bool       `json:"enabled"`
	Status              string     `json:"status"`
	Mode                string     `json:"mode"`
	Health              string     `json:"health"`
	LastSyncAt          *time.Time `json:"lastSyncAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	HasAccessToken      bool       `json:"hasAccessToken"`
	HasPassword         bool       `json:"hasPassword"`
	AccessToken         string     `json:"-"`
	Password            string     `json:"-"`
}

type IntegrationLog struct {
	ID            string    `json:"id"`
	IntegrationID string    `json:"integrationId"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	CreatedAt     time.Time `json:"createdAt"`
}

type IntegrationHealth struct {
	Status     string            `json:"status"`
	Healthy    bool              `json:"healthy"`
	Message    string            `json:"message"`
	CheckedAt  time.Time         `json:"checkedAt"`
	LatencyMs  int               `json:"latencyMs"`
	Attributes map[string]string `json:"attributes"`
}

type IntegrationSyncResult struct {
	Status    string            `json:"status"`
	Message   string            `json:"message"`
	SyncedAt  time.Time         `json:"syncedAt"`
	Resources map[string]int    `json:"resources"`
	Metadata  map[string]string `json:"metadata"`
}

type CreateIntegrationRequest struct {
	Name                string `json:"name" binding:"required,min=2,max=80"`
	Type                string `json:"type" binding:"required,oneof=Prometheus Grafana Harbor Vault GitHub GitHubActions ArgoCD Kubernetes Loki Tempo SonarQube Trivy Semgrep Gitleaks Falco Wazuh Keycloak OPA-Gatekeeper"`
	Category            string `json:"category" binding:"omitempty,max=80"`
	EndpointURL         string `json:"endpointUrl" binding:"required,url"`
	AccessToken         string `json:"accessToken" binding:"omitempty"`
	Username            string `json:"username" binding:"omitempty,max=120"`
	Password            string `json:"password" binding:"omitempty"`
	Namespace           string `json:"namespace" binding:"omitempty,max=120"`
	TLSVerify           *bool  `json:"tlsVerify"`
	SyncIntervalSeconds int    `json:"syncIntervalSeconds" binding:"omitempty,min=15,max=86400"`
	Enabled             bool   `json:"enabled"`
}

type UpdateIntegrationRequest struct {
	Name                *string `json:"name" binding:"omitempty,min=2,max=80"`
	Category            *string `json:"category" binding:"omitempty,max=80"`
	EndpointURL         *string `json:"endpointUrl" binding:"omitempty,url"`
	AccessToken         *string `json:"accessToken"`
	Username            *string `json:"username" binding:"omitempty,max=120"`
	Password            *string `json:"password"`
	Namespace           *string `json:"namespace" binding:"omitempty,max=120"`
	TLSVerify           *bool   `json:"tlsVerify"`
	SyncIntervalSeconds *int    `json:"syncIntervalSeconds" binding:"omitempty,min=15,max=86400"`
	Enabled             *bool   `json:"enabled"`
}

type SetIntegrationEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Environment string    `json:"environment"`
	IsSensitive bool      `json:"isSensitive"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateSettingRequest struct {
	Value       string `json:"value" binding:"required"`
	Environment string `json:"environment" binding:"omitempty,max=40"`
	IsSensitive bool   `json:"isSensitive"`
}

type IntegrationStatusView struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Health        string     `json:"health"`
	LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
	LastCheckedAt *time.Time `json:"lastCheckedAt,omitempty"`
	Enabled       bool       `json:"enabled"`
}

type ConnectionHistory struct {
	ID            string    `json:"id"`
	IntegrationID string    `json:"integrationId"`
	Action        string    `json:"action"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	LatencyMs     int       `json:"latencyMs"`
	CheckedAt     time.Time `json:"checkedAt"`
}

type IntegrationType struct {
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
}
