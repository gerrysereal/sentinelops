'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { RegistryArtifact } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

export default function RegistryPage() {
  const { data, isLoading, error } = useQuery({ queryKey: ['registry-artifacts'], queryFn: api.registryArtifacts });

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load registry artifacts." />;

  const columns: Column<RegistryArtifact>[] = [
    { key: 'name', header: 'Artifact', cell: (row) => <span className="font-medium text-slate-100">{row.name}</span> },
    { key: 'registry', header: 'Registry', cell: (row) => row.registry },
    { key: 'image', header: 'Image', cell: (row) => <code className="break-all text-cyan-200">{row.image}</code> },
    { key: 'sbom', header: 'SBOM', cell: (row) => <Badge variant="muted">{row.sbom}</Badge> },
    { key: 'signature', header: 'Signature', cell: (row) => <Badge variant="success">{row.signature}</Badge> },
    { key: 'scan', header: 'Scan', cell: (row) => <Badge variant={row.scan === 'passed' ? 'success' : 'danger'}>{row.scan}</Badge> }
  ];

  return <DataTable columns={columns} data={data} />;
}
