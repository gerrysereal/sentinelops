import type {
  Application,
  Deployment,
  Integration,
  ObservabilitySignal,
  Overview,
  PipelineRun,
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

  return response.json() as Promise<T>;
}

export const api = {
  overview: () => request<Overview>('/overview'),
  applications: () => request<Application[]>('/applications'),
  pipelines: () => request<PipelineRun[]>('/pipelines'),
  deployments: () => request<Deployment[]>('/deployments'),
  alerts: () => request<SecurityAlert[]>('/security/alerts'),
  signals: () => request<ObservabilitySignal[]>('/observability/signals'),
  integrations: () => request<Integration[]>('/integrations')
};
