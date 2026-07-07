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

type Integration struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Endpoint string `json:"endpoint"`
}

type Overview struct {
	ApplicationsCount int             `json:"applicationsCount"`
	ClustersCount     int             `json:"clustersCount"`
	NodesCount        int             `json:"nodesCount"`
	PodsCount         int             `json:"podsCount"`
	DeploymentStatus  map[string]int  `json:"deploymentStatus"`
	PipelineStatus    map[string]int  `json:"pipelineStatus"`
	SecuritySeverity  map[string]int  `json:"securitySeverity"`
	ResourceUsage     map[string]int  `json:"resourceUsage"`
	RecentAlerts      []SecurityAlert `json:"recentAlerts"`
	Integrations      []Integration   `json:"integrations"`
}

type CreateApplicationRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=80"`
	Owner       string `json:"owner" binding:"required,min=2,max=80"`
	Repository  string `json:"repository" binding:"required"`
	Environment string `json:"environment" binding:"required,oneof=dev staging production"`
}
