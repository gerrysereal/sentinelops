'use client';

import { FormEvent, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { SecurityAlert } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

const inputClass = 'rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50';

function severityVariant(severity: string): 'success' | 'warning' | 'danger' | 'muted' | 'default' {
  if (severity === 'critical' || severity === 'high') return 'danger';
  if (severity === 'medium') return 'warning';
  if (severity === 'low') return 'muted';
  return 'default';
}

export default function SecurityPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState({ application: '', source: 'Trivy', severity: 'medium', title: '' });
  const [selectedAlert, setSelectedAlert] = useState<SecurityAlert | null>(null);
  const alerts = useQuery({ queryKey: ['security-alerts'], queryFn: api.alerts });
  const apps = useQuery({ queryKey: ['applications'], queryFn: api.applications });

  const scanMutation = useMutation({
    mutationFn: api.runSecurityScan,
    onSuccess: async () => {
      setForm((current) => ({ ...current, title: '' }));
      await invalidateSecurity();
    }
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'triaged' | 'resolved' }) => api.updateAlertStatus(id, status),
    onSuccess: async () => invalidateSecurity()
  });

  async function invalidateSecurity() {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['security-alerts'] }),
      queryClient.invalidateQueries({ queryKey: ['registry-artifacts'] }),
      queryClient.invalidateQueries({ queryKey: ['overview'] })
    ]);
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const application = form.application || apps.data?.[0]?.name;
    if (!application) return;
    scanMutation.mutate({ ...form, application });
  }

  if (alerts.isLoading || apps.isLoading) return <LoadingState />;
  if (alerts.error || !alerts.data) return <ErrorState message="Unable to load security alerts." />;

  const columns: Column<SecurityAlert>[] = [
    { key: 'source', header: 'Source', cell: (row) => row.source },
    { key: 'title', header: 'Finding', cell: (row) => <span className="font-medium text-slate-100">{row.title}</span> },
    { key: 'app', header: 'Application', cell: (row) => row.application },
    { key: 'severity', header: 'Severity', cell: (row) => <Badge variant={severityVariant(row.severity)}>{row.severity}</Badge> },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={row.status === 'resolved' ? 'success' : 'muted'}>{row.status}</Badge> },
    {
      key: 'actions',
      header: 'Actions',
      cell: (row) => (
        <div className="flex gap-2">
          <Button className="px-3 py-1 text-xs" disabled={statusMutation.isPending || row.status === 'triaged'} onClick={() => statusMutation.mutate({ id: row.id, status: 'triaged' })}>Triage</Button>
          <Button className="px-3 py-1 text-xs" disabled={statusMutation.isPending || row.status === 'resolved'} onClick={() => statusMutation.mutate({ id: row.id, status: 'resolved' })}>Resolve</Button>
        </div>
      )
    }
  ];

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader><CardTitle>Run Security Scan</CardTitle></CardHeader>
        <CardContent>
          <form onSubmit={submit} className="grid gap-3 md:grid-cols-5">
            <select className={inputClass} value={form.application} onChange={(e) => setForm({ ...form, application: e.target.value })}>
              {(apps.data ?? []).map((app) => <option key={app.id} value={app.name}>{app.name}</option>)}
            </select>
            <select className={inputClass} value={form.source} onChange={(e) => setForm({ ...form, source: e.target.value })}>
              {['Trivy', 'Semgrep', 'Gitleaks', 'Falco', 'Wazuh', 'OPA-Gatekeeper'].map((tool) => <option key={tool} value={tool}>{tool}</option>)}
            </select>
            <select className={inputClass} value={form.severity} onChange={(e) => setForm({ ...form, severity: e.target.value })}>
              {['critical', 'high', 'medium', 'low'].map((severity) => <option key={severity} value={severity}>{severity}</option>)}
            </select>
            <input className={inputClass} placeholder="optional finding title" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} />
            <Button disabled={scanMutation.isPending}>{scanMutation.isPending ? 'Scanning...' : 'Run Scan'}</Button>
          </form>
          {scanMutation.error ? <p className="mt-3 text-xs text-rose-300">{scanMutation.error.message}</p> : null}
        </CardContent>
      </Card>
      <DataTable columns={columns} data={alerts.data} onRowClick={setSelectedAlert} />
      {selectedAlert ? (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between gap-3">
              <CardTitle>Security Finding Detail</CardTitle>
              <Button className="border-slate-700 bg-slate-900 text-slate-200 hover:bg-slate-800" onClick={() => setSelectedAlert(null)}>Close</Button>
            </div>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-2">
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Finding</p><p className="text-slate-100">{selectedAlert.title}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Application</p><p className="text-slate-100">{selectedAlert.application}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Tool</p><p className="text-slate-100">{selectedAlert.source}</p></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Severity / Status</p><div className="mt-2 flex gap-2"><Badge variant={severityVariant(selectedAlert.severity)}>{selectedAlert.severity}</Badge><Badge variant={selectedAlert.status === 'resolved' ? 'success' : 'muted'}>{selectedAlert.status}</Badge></div></div>
            <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 md:col-span-2"><p className="text-xs text-slate-500">Recommended action</p><p className="mt-2 text-sm text-slate-300">Triage untuk acknowledge finding. Resolve jika sudah ada fix, exception, atau false positive yang terdokumentasi.</p></div>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
