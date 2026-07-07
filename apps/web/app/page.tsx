'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { AlertsCard } from '@/components/dashboard/alerts-card';
import { ErrorState } from '@/components/dashboard/error-state';
import { IntegrationsCard } from '@/components/dashboard/integrations-card';
import { LoadingState } from '@/components/dashboard/loading-state';
import { MetricCard } from '@/components/dashboard/metric-card';
import { ProgressCard } from '@/components/dashboard/progress-card';
import { StatusList } from '@/components/dashboard/status-list';

export default function DashboardPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['overview'], queryFn: api.overview });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load SentinelOps overview. Check whether the API is running on port 8080." />;

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="Applications" value={data.applicationsCount} hint="Registered services" />
        <MetricCard label="Clusters" value={data.clustersCount} hint="Managed GitOps targets" />
        <MetricCard label="Nodes" value={data.nodesCount} hint="k3s worker/control nodes" />
        <MetricCard label="Pods" value={data.podsCount} hint="Runtime workloads" />
      </section>

      <section className="grid gap-4 xl:grid-cols-3">
        <StatusList title="Deployment Status" data={data.deploymentStatus} />
        <StatusList title="Pipeline Status" data={data.pipelineStatus} />
        <StatusList title="Security Severity" data={data.securitySeverity} />
      </section>

      <section className="grid gap-4 xl:grid-cols-[1fr_1.3fr]">
        <ProgressCard data={data.resourceUsage} />
        <AlertsCard alerts={data.recentAlerts} />
      </section>

      <IntegrationsCard integrations={data.integrations} />
    </div>
  );
}
