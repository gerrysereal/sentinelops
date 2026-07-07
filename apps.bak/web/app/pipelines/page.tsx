'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { PipelineRun } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (status === 'success') return 'success';
  if (status === 'failed') return 'danger';
  if (status === 'running') return 'default';
  return 'muted';
}

export default function PipelinesPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['pipelines'], queryFn: api.pipelines });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load CI/CD pipelines." />;

  const columns: Column<PipelineRun>[] = [
    { key: 'app', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.applicationName}</span> },
    { key: 'branch', header: 'Branch', cell: (row) => row.branch },
    { key: 'commit', header: 'Commit', cell: (row) => <code className="text-cyan-200">{row.commitSha}</code> },
    { key: 'stage', header: 'Stage', cell: (row) => row.stage },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
    { key: 'duration', header: 'Duration', cell: (row) => `${row.durationSeconds}s` }
  ];

  return <DataTable columns={columns} data={data} />;
}
