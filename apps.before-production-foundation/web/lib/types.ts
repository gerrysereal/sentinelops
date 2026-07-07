export type Application = {
  id: string;
  name: string;
  owner: string;
  repository: string;
  environment: string;
  status: string;
  createdAt: string;
};

export type PipelineRun = {
  id: string;
  applicationId: string;
  applicationName: string;
  branch: string;
  commitSha: string;
  status: string;
  stage: string;
  durationSeconds: number;
  finishedAt: string;
};

export type Deployment = {
  id: string;
  applicationId: string;
  applicationName: string;
  cluster: string;
  namespace: string;
  image: string;
  version: string;
  syncStatus: string;
  healthStatus: string;
  deployedAt: string;
};

export type SecurityAlert = {
  id: string;
  source: string;
  severity: string;
  title: string;
  application: string;
  status: string;
  detectedAt: string;
};

export type ObservabilitySignal = {
  id: string;
  source: string;
  type: string;
  status: string;
  message: string;
  createdAt: string;
};

export type Integration = {
  id: string;
  name: string;
  type: string;
  category: string;
  endpointUrl: string;
  username?: string;
  namespace?: string;
  tlsVerify: boolean;
  syncIntervalSeconds: number;
  enabled: boolean;
  status: string;
  mode: string;
  health: string;
  lastSyncAt?: string;
  createdAt: string;
  updatedAt: string;
  hasAccessToken: boolean;
  hasPassword: boolean;
};

export type IntegrationLog = {
  id: string;
  integrationId: string;
  action: string;
  status: string;
  message: string;
  createdAt: string;
};

export type IntegrationHealth = {
  status: string;
  healthy: boolean;
  message: string;
  checkedAt: string;
  attributes?: Record<string, string>;
};

export type IntegrationSyncResult = {
  status: string;
  message: string;
  syncedAt: string;
  resources?: Record<string, number>;
  metadata?: Record<string, string>;
};

export type RegistryArtifact = {
  id: string;
  name: string;
  registry: string;
  image: string;
  version: string;
  sbom: string;
  signature: string;
  scan: string;
};

export type Overview = {
  applicationsCount: number;
  clustersCount: number;
  nodesCount: number;
  podsCount: number;
  deploymentStatus: Record<string, number>;
  pipelineStatus: Record<string, number>;
  securitySeverity: Record<string, number>;
  resourceUsage: Record<string, number>;
  recentAlerts: SecurityAlert[];
  integrations: Integration[];
};
