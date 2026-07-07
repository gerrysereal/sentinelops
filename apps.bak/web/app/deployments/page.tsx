'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Deployment } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

function healthVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (status === 'healthy') return 'success';
  if (status === 'degraded') return 'warning';
  if (status === 'progressing') return 'default';
  return 'muted';
}

export default function DeploymentsPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['deployments'], queryFn: api.deployments });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load deployments." />;

  const columns: Column<Deployment>[] = [
    { key: 'app', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.applicationName}</span> },
    { key: 'cluster', header: 'Cluster', cell: (row) => row.cluster },
    { key: 'namespace', header: 'Namespace', cell: (row) => row.namespace },
    { key: 'version', header: 'Version', cell: (row) => row.version },
    { key: 'sync', header: 'Sync', cell: (row) => <Badge variant={row.syncStatus === 'synced' ? 'success' : 'warning'}>{row.syncStatus}</Badge> },
    { key: 'health', header: 'Health', cell: (row) => <Badge variant={healthVariant(row.healthStatus)}>{row.healthStatus}</Badge> }
  ];

  return <DataTable columns={columns} data={data} />;
}
