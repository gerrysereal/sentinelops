'use client';

import { FormEvent, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Deployment } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

const inputClass = 'rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50';

function healthVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (status === 'healthy') return 'success';
  if (status === 'degraded') return 'warning';
  if (status === 'progressing') return 'default';
  return 'muted';
}

export default function DeploymentsPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState({ applicationId: '', cluster: 'k3s-prod-01', namespace: '', version: 'v1.0.0' });
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const deployments = useQuery({ queryKey: ['deployments'], queryFn: api.deployments });
  const apps = useQuery({ queryKey: ['applications'], queryFn: api.applications });

  const deployMutation = useMutation({
    mutationFn: api.createDeployment,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['deployments'] }),
        queryClient.invalidateQueries({ queryKey: ['registry-artifacts'] }),
        queryClient.invalidateQueries({ queryKey: ['observability-signals'] }),
        queryClient.invalidateQueries({ queryKey: ['overview'] })
      ]);
    }
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const applicationId = form.applicationId || apps.data?.[0]?.id;
    if (!applicationId) return;
    deployMutation.mutate({ ...form, applicationId });
  }

  if (deployments.isLoading || apps.isLoading) return <LoadingState />;
  if (deployments.error || !deployments.data) return <ErrorState message="Unable to load deployments." />;

  const columns: Column<Deployment>[] = [
    { key: 'app', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.applicationName}</span> },
    { key: 'cluster', header: 'Cluster', cell: (row) => row.cluster },
    { key: 'namespace', header: 'Namespace', cell: (row) => row.namespace },
    { key: 'version', header: 'Version', cell: (row) => row.version },
    { key: 'sync', header: 'Sync', cell: (row) => <Badge variant={row.syncStatus === 'synced' ? 'success' : 'warning'}>{row.syncStatus}</Badge> },
    { key: 'health', header: 'Health', cell: (row) => <Badge variant={healthVariant(row.healthStatus)}>{row.healthStatus}</Badge> }
  ];

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader><CardTitle>GitOps Deploy</CardTitle></CardHeader>
        <CardContent>
          <form onSubmit={submit} className="grid gap-3 md:grid-cols-5">
            <select className={inputClass} value={form.applicationId} onChange={(e) => setForm({ ...form, applicationId: e.target.value })}>
              {(apps.data ?? []).map((app) => <option key={app.id} value={app.id}>{app.name}</option>)}
            </select>
            <input className={inputClass} value={form.cluster} onChange={(e) => setForm({ ...form, cluster: e.target.value })} />
            <input className={inputClass} placeholder="namespace" value={form.namespace} onChange={(e) => setForm({ ...form, namespace: e.target.value })} />
            <input className={inputClass} placeholder="version" value={form.version} onChange={(e) => setForm({ ...form, version: e.target.value })} />
            <Button disabled={deployMutation.isPending}>{deployMutation.isPending ? 'Deploying...' : 'Deploy'}</Button>
          </form>
          {deployMutation.error ? <p className="mt-3 text-xs text-rose-300">{deployMutation.error.message}</p> : null}
        </CardContent>
      </Card>
      <DataTable columns={columns} data={deployments.data} onRowClick={setSelectedDeployment} />
      {selectedDeployment ? (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between gap-3">
              <CardTitle>Deployment Detail</CardTitle>
              <Button className="border-slate-700 bg-slate-900 text-slate-200 hover:bg-slate-800" onClick={() => setSelectedDeployment(null)}>Close</Button>
            </div>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-3">
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Application</p><p className="text-slate-100">{selectedDeployment.applicationName}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Cluster / Namespace</p><p className="text-slate-100">{selectedDeployment.cluster} · {selectedDeployment.namespace}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Health</p><div className="mt-2"><Badge variant={healthVariant(selectedDeployment.healthStatus)}>{selectedDeployment.healthStatus}</Badge></div></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 md:col-span-3"><p className="text-xs text-slate-500">Image</p><p className="mt-2 break-all text-sm text-slate-300">{selectedDeployment.image}:{selectedDeployment.version}</p></div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
