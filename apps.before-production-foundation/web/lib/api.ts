import type {
  Application,
  Deployment,
  Integration,
  IntegrationHealth,
  IntegrationLog,
  IntegrationSyncResult,
  ObservabilitySignal,
  Overview,
  PipelineRun,
  RegistryArtifact,
  SecurityAlert
} from './types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/api/v1';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {})
    },
    cache: 'no-store'
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed: ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

function post<T>(path: string, body?: unknown): Promise<T> {
  return request<T>(path, { method: 'POST', body: body === undefined ? undefined : JSON.stringify(body) });
}

function put<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PUT', body: JSON.stringify(body) });
}

function patch<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PATCH', body: JSON.stringify(body) });
}

function del(path: string): Promise<void> {
  return request<void>(path, { method: 'DELETE' });
}

export type IntegrationPayload = {
  name: string;
  type: string;
  category?: string;
  endpointUrl: string;
  accessToken?: string;
  username?: string;
  password?: string;
  namespace?: string;
  tlsVerify: boolean;
  syncIntervalSeconds: number;
  enabled: boolean;
};

export const api = {
  overview: () => request<Overview>('/overview'),
  applications: () => request<Application[]>('/applications'),
  createApplication: (body: Pick<Application, 'name' | 'owner' | 'repository' | 'environment'>) => post<Application>('/applications', body),
  pipelines: () => request<PipelineRun[]>('/pipelines'),
  runPipeline: (body: { applicationId: string; branch?: string; stage?: string; status?: string }) => post<PipelineRun>('/pipelines/run', body),
  deployments: () => request<Deployment[]>('/deployments'),
  createDeployment: (body: { applicationId: string; cluster?: string; namespace?: string; version?: string }) => post<Deployment>('/deployments', body),
  alerts: () => request<SecurityAlert[]>('/security/alerts'),
  runSecurityScan: (body: { application: string; source?: string; severity?: string; title?: string }) => post<SecurityAlert>('/security/scans', body),
  updateAlertStatus: (id: string, status: 'open' | 'triaged' | 'resolved') => patch<SecurityAlert>(`/security/alerts/${id}/status`, { status }),
  signals: () => request<ObservabilitySignal[]>('/observability/signals'),
  integrations: () => request<Integration[]>('/integrations'),
  createIntegration: (body: IntegrationPayload) => post<Integration>('/integrations', body),
  updateIntegration: (id: string, body: Partial<IntegrationPayload>) => put<Integration>(`/integrations/${id}`, body),
  deleteIntegration: (id: string) => del(`/integrations/${id}`),
  setIntegrationEnabled: (id: string, enabled: boolean) => patch<Integration>(`/integrations/${id}/enabled`, { enabled }),
  testIntegration: (id: string) => post<IntegrationHealth>(`/integrations/${id}/test`),
  syncIntegration: (id: string) => post<IntegrationSyncResult>(`/integrations/${id}/sync`),
  integrationLogs: (id: string) => request<IntegrationLog[]>(`/integrations/${id}/logs`),
  registryArtifacts: () => request<RegistryArtifact[]>('/registry/artifacts')
};
