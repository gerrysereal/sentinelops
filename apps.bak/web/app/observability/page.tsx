'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { ObservabilitySignal } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (status === 'healthy') return 'success';
  if (status === 'warning') return 'warning';
  return 'muted';
}

export default function ObservabilityPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['observability-signals'], queryFn: api.signals });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load observability signals." />;

  const columns: Column<ObservabilitySignal>[] = [
    { key: 'source', header: 'Source', cell: (row) => row.source },
    { key: 'type', header: 'Signal', cell: (row) => <Badge variant="default">{row.type}</Badge> },
    { key: 'message', header: 'Message', cell: (row) => <span className="font-medium text-slate-100">{row.message}</span> },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> }
  ];

  return <DataTable columns={columns} data={data} />;
}
