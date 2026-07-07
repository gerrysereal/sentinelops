'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { SecurityAlert } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

function severityVariant(severity: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (severity === 'high') return 'danger';
  if (severity === 'medium') return 'warning';
  if (severity === 'low') return 'muted';
  return 'default';
}

export default function SecurityPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['security-alerts'], queryFn: api.alerts });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load security alerts." />;

  const columns: Column<SecurityAlert>[] = [
    { key: 'source', header: 'Source', cell: (row) => row.source },
    { key: 'title', header: 'Finding', cell: (row) => <span className="font-medium text-slate-100">{row.title}</span> },
    { key: 'app', header: 'Application', cell: (row) => row.application },
    { key: 'severity', header: 'Severity', cell: (row) => <Badge variant={severityVariant(row.severity)}>{row.severity}</Badge> },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant="muted">{row.status}</Badge> }
  ];

  return <DataTable columns={columns} data={data} />;
}
