package domain

import "time"

type Application struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Owner       string    `json:"owner"`
	Repository  string    `json:"repository"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PipelineRun struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"applicationId"`
	ApplicationName string    `json:"applicationName"`
	Branch          string    `json:"branch"`
	CommitSHA       string    `json:"commitSha"`
	Status          string    `json:"status"`
	Stage           string    `json:"stage"`
	DurationSeconds int       `json:"durationSeconds"`
	FinishedAt      time.Time `json:"finishedAt"`
}

type Deployment struct {
	ID              string    `json:"id"`
	ApplicationID   string    `json:"applicationId"`
	ApplicationName string    `json:"applicationName"`
	Cluster         string    `json:"cluster"`
	Namespace       string    `json:"namespace"`
	Image           string    `json:"image"`
	Version         string    `json:"version"`
	SyncStatus      string    `json:"syncStatus"`
	HealthStatus    string    `json:"healthStatus"`
	DeployedAt      time.Time `json:"deployedAt"`
}

type SecurityAlert struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Application string    `json:"application"`
	Status      string    `json:"status"`
	DetectedAt  time.Time `json:"detectedAt"`
}

type ObservabilitySignal struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegistryArtifact struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Registry  string `json:"registry"`
	Image     string `json:"image"`
	Version   string `json:"version"`
	SBOM      string `json:"sbom"`
	Signature string `json:"signature"`
	Scan      string `json:"scan"`
}

type Overview struct {
	ApplicationsCount int                 `json:"applicationsCount"`
	ClustersCount     int                 `json:"clustersCount"`
	NodesCount        int                 `json:"nodesCount"`
	PodsCount         int                 `json:"podsCount"`
	DeploymentStatus  map[string]int      `json:"deploymentStatus"`
	PipelineStatus    map[string]int      `json:"pipelineStatus"`
	SecuritySeverity  map[string]int      `json:"securitySeverity"`
	ResourceUsage     map[string]int      `json:"resourceUsage"`
	RecentAlerts      []SecurityAlert     `json:"recentAlerts"`
	Integrations      []IntegrationConfig `json:"integrations"`
}

type CreateApplicationRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=80"`
	Owner       string `json:"owner" binding:"required,min=2,max=80"`
	Repository  string `json:"repository" binding:"required"`
	Environment string `json:"environment" binding:"required,oneof=dev staging production"`
}

type CreatePipelineRunRequest struct {
	ApplicationID string `json:"applicationId" binding:"required"`
	Branch        string `json:"branch" binding:"omitempty,min=1,max=80"`
	Stage         string `json:"stage" binding:"omitempty,min=2,max=80"`
	Status        string `json:"status" binding:"omitempty,oneof=success failed running pending"`
}

type CreateDeploymentRequest struct {
	ApplicationID string `json:"applicationId" binding:"required"`
	Cluster       string `json:"cluster" binding:"omitempty,max=80"`
	Namespace     string `json:"namespace" binding:"omitempty,max=80"`
	Version       string `json:"version" binding:"omitempty,max=40"`
}

type CreateSecurityScanRequest struct {
	Application string `json:"application" binding:"required"`
	Source      string `json:"source" binding:"omitempty,oneof=Trivy Semgrep Gitleaks Falco Wazuh OPA-Gatekeeper"`
	Severity    string `json:"severity" binding:"omitempty,oneof=critical high medium low"`
	Title       string `json:"title" binding:"omitempty,max=160"`
}

type UpdateAlertStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open triaged resolved"`
}
