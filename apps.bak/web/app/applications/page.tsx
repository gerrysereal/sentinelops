'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Application } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

export default function ApplicationsPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['applications'], queryFn: api.applications });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load applications." />;

  const columns: Column<Application>[] = [
    { key: 'name', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.name}</span> },
    { key: 'owner', header: 'Owner', cell: (row) => row.owner },
    { key: 'env', header: 'Environment', cell: (row) => <Badge variant="muted">{row.environment}</Badge> },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={row.status === 'healthy' ? 'success' : 'warning'}>{row.status}</Badge> },
    { key: 'repo', header: 'Repository', cell: (row) => <span className="text-slate-500">{row.repository}</span> }
  ];

  return <DataTable columns={columns} data={data} />;
}
