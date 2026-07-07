'use client';

import { FormEvent, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { PipelineRun } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

const inputClass = 'rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50';

function statusVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (status === 'success') return 'success';
  if (status === 'failed') return 'danger';
  if (status === 'running') return 'default';
  return 'muted';
}

export default function PipelinesPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState({ applicationId: '', branch: 'main', stage: 'trivy-image-scan', status: 'success' });
  const [selectedRun, setSelectedRun] = useState<PipelineRun | null>(null);
  const pipelines = useQuery({ queryKey: ['pipelines'], queryFn: api.pipelines });
  const apps = useQuery({ queryKey: ['applications'], queryFn: api.applications });

  const runMutation = useMutation({
    mutationFn: api.runPipeline,
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['pipelines'] }),
        queryClient.invalidateQueries({ queryKey: ['observability-signals'] }),
        queryClient.invalidateQueries({ queryKey: ['overview'] })
      ]);
    }
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const applicationId = form.applicationId || apps.data?.[0]?.id;
    if (!applicationId) return;
    runMutation.mutate({ ...form, applicationId });
  }

  if (pipelines.isLoading || apps.isLoading) return <LoadingState />;
  if (pipelines.error || !pipelines.data) return <ErrorState message="Unable to load CI/CD pipelines." />;

  const columns: Column<PipelineRun>[] = [
    { key: 'app', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.applicationName}</span> },
    { key: 'branch', header: 'Branch', cell: (row) => row.branch },
    { key: 'commit', header: 'Commit', cell: (row) => <code className="text-cyan-200">{row.commitSha}</code> },
    { key: 'stage', header: 'Stage', cell: (row) => row.stage },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={statusVariant(row.status)}>{row.status}</Badge> },
    { key: 'duration', header: 'Duration', cell: (row) => row.durationSeconds ? `${row.durationSeconds}s` : 'live' }
  ];

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader><CardTitle>Run Pipeline</CardTitle></CardHeader>
        <CardContent>
          <form onSubmit={submit} className="grid gap-3 md:grid-cols-5">
            <select className={inputClass} value={form.applicationId} onChange={(e) => setForm({ ...form, applicationId: e.target.value })}>
              {(apps.data ?? []).map((app) => <option key={app.id} value={app.id}>{app.name}</option>)}
            </select>
            <input className={inputClass} value={form.branch} onChange={(e) => setForm({ ...form, branch: e.target.value })} />
            <select className={inputClass} value={form.stage} onChange={(e) => setForm({ ...form, stage: e.target.value })}>
              <option value="semgrep-sast">semgrep-sast</option>
              <option value="trivy-image-scan">trivy-image-scan</option>
              <option value="push-image">push-image</option>
              <option value="argocd-sync">argocd-sync</option>
            </select>
            <select className={inputClass} value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option value="success">success</option>
              <option value="failed">failed</option>
              <option value="running">running</option>
              <option value="pending">pending</option>
            </select>
            <Button disabled={runMutation.isPending}>{runMutation.isPending ? 'Running...' : 'Run Pipeline'}</Button>
          </form>
          {runMutation.error ? <p className="mt-3 text-xs text-rose-300">{runMutation.error.message}</p> : null}
        </CardContent>
      </Card>
      <DataTable columns={columns} data={pipelines.data} onRowClick={setSelectedRun} />
      {selectedRun ? (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between gap-3">
              <CardTitle>Pipeline Run Detail</CardTitle>
              <Button className="border-slate-700 bg-slate-900 text-slate-200 hover:bg-slate-800" onClick={() => setSelectedRun(null)}>Close</Button>
            </div>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-3">
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Application</p><p className="text-slate-100">{selectedRun.applicationName}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Branch / Commit</p><p className="text-slate-100">{selectedRun.branch} · {selectedRun.commitSha}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Status</p><div className="mt-2"><Badge variant={statusVariant(selectedRun.status)}>{selectedRun.status}</Badge></div></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 md:col-span-3"><p className="text-xs text-slate-500">Stage</p><p className="mt-2 text-sm text-slate-300">{selectedRun.stage} finished in {selectedRun.durationSeconds || 'live'} seconds.</p></div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
