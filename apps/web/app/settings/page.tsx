'use client';

import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, type IntegrationPayload } from '@/lib/api';
import type { ConnectionHistory, Integration, IntegrationLog, IntegrationType } from '@/lib/types';
import { StatusDot } from '@/components/dashboard/status-dot';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const emptyForm: IntegrationPayload = {
  name: '',
  type: 'Prometheus',
  category: '',
  endpointUrl: '',
  accessToken: '',
  username: '',
  password: '',
  namespace: '',
  tlsVerify: true,
  syncIntervalSeconds: 60,
  enabled: true
};

function badgeVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' {
  if (status === 'connected' || status === 'healthy') return 'success';
  if (status === 'error' || status === 'disconnected') return 'danger';
  if (status === 'disabled') return 'muted';
  return 'warning';
}

function toForm(item: Integration): IntegrationPayload {
  return {
    name: item.name,
    type: item.type,
    category: item.category,
    endpointUrl: item.endpointUrl,
    accessToken: '',
    username: item.username ?? '',
    password: '',
    namespace: item.namespace ?? '',
    tlsVerify: item.tlsVerify,
    syncIntervalSeconds: item.syncIntervalSeconds,
    enabled: item.enabled
  };
}

export default function SettingsPage() {
  const queryClient = useQueryClient();
  const [selected, setSelected] = useState<Integration | null>(null);
  const [form, setForm] = useState<IntegrationPayload>(emptyForm);
  const [feedback, setFeedback] = useState<string>('');

  const integrations = useQuery({ queryKey: ['integrations'], queryFn: api.integrations });
  const integrationTypes = useQuery({ queryKey: ['integration-types'], queryFn: api.integrationTypes });
  const logs = useQuery({
    queryKey: ['integration-logs', selected?.id],
    queryFn: () => api.integrationLogs(selected!.id),
    enabled: Boolean(selected?.id)
  });
  const history = useQuery({
    queryKey: ['connection-history', selected?.id],
    queryFn: () => api.connectionHistory(selected!.id),
    enabled: Boolean(selected?.id)
  });

  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['integrations'] }),
      queryClient.invalidateQueries({ queryKey: ['overview'] }),
      queryClient.invalidateQueries({ queryKey: ['integration-logs'] }),
      queryClient.invalidateQueries({ queryKey: ['connection-history'] })
    ]);
  };

  const createMutation = useMutation({
    mutationFn: api.createIntegration,
    onSuccess: async (created) => {
      setFeedback(`Saved ${created.name}`);
      setSelected(created);
      setForm(toForm(created));
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: Partial<IntegrationPayload> }) => api.updateIntegration(id, body),
    onSuccess: async (updated) => {
      setFeedback(`Updated ${updated.name}`);
      setSelected(updated);
      setForm(toForm(updated));
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const deleteMutation = useMutation({
    mutationFn: api.deleteIntegration,
    onSuccess: async () => {
      setFeedback('Integration deleted');
      setSelected(null);
      setForm(emptyForm);
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const enableMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) => api.setIntegrationEnabled(id, enabled),
    onSuccess: async (updated) => {
      setFeedback(`${updated.name} is now ${updated.enabled ? 'enabled' : 'disabled'}`);
      setSelected(updated);
      setForm(toForm(updated));
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const testMutation = useMutation({
    mutationFn: api.testIntegration,
    onSuccess: async (result) => {
      setFeedback(`Connection test: ${result.status} — ${result.message}`);
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const syncMutation = useMutation({
    mutationFn: api.syncIntegration,
    onSuccess: async (result) => {
      setFeedback(`Sync complete: ${result.status} — ${result.message}`);
      await invalidate();
    },
    onError: (error: Error) => setFeedback(error.message)
  });

  const isBusy = createMutation.isPending || updateMutation.isPending || deleteMutation.isPending || enableMutation.isPending || testMutation.isPending || syncMutation.isPending;

  const integrationList = useMemo(() => integrations.data ?? [], [integrations.data]);
  const typeList = useMemo(() => integrationTypes.data ?? [], [integrationTypes.data]);

  if (integrations.isLoading || integrationTypes.isLoading) return <LoadingState label="Loading platform settings..." />;
  if (integrations.error || integrationTypes.error) return <ErrorState message="Unable to load integration settings from the SentinelOps API." />;

  const submit = () => {
    setFeedback('');
    const body: IntegrationPayload = {
      ...form,
      syncIntervalSeconds: Number(form.syncIntervalSeconds || 60)
    };
    if (selected) {
      const payload: Partial<IntegrationPayload> = { ...body };
      if (!payload.accessToken) delete payload.accessToken;
      if (!payload.password) delete payload.password;
      updateMutation.mutate({ id: selected.id, body: payload });
    } else {
      createMutation.mutate(body);
    }
  };

  const selectIntegration = (item: Integration) => {
    setSelected(item);
    setForm(toForm(item));
    setFeedback('');
  };

  const resetForm = () => {
    setSelected(null);
    setForm({ ...emptyForm, type: typeList[0]?.type ?? emptyForm.type, category: typeList[0]?.category ?? '' });
    setFeedback('');
  };

  return (
    <div className="space-y-6">
      <section className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle>Integrations</CardTitle>
            <Badge variant="muted">stored in PostgreSQL</Badge>
          </CardHeader>
          <CardContent className="space-y-3">
            {integrationList.map((item) => (
              <button
                key={item.id}
                type="button"
                onClick={() => selectIntegration(item)}
                className="w-full rounded-xl border border-slate-800 bg-slate-900/70 p-4 text-left transition hover:border-cyan-400/40 hover:bg-slate-900"
              >
                <div className="mb-2 flex items-start justify-between gap-4">
                  <div>
                    <div className="flex items-center gap-2 font-medium text-slate-100">
                      <StatusDot status={item.status} />
                      {item.name}
                    </div>
                    <div className="mt-1 text-xs text-slate-500">{item.type} • {item.category} • {item.mode}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={badgeVariant(item.status)}>{item.status}</Badge>
                    <Badge variant={item.enabled ? 'success' : 'muted'}>{item.enabled ? 'enabled' : 'disabled'}</Badge>
                  </div>
                </div>
                <div className="truncate text-xs text-slate-400">{item.endpointUrl}</div>
                <div className="mt-2 text-xs text-slate-500">
                  Last sync: {item.lastSyncAt ? new Date(item.lastSyncAt).toLocaleString() : 'never'} • Token: {item.hasAccessToken ? 'encrypted' : 'not set'}
                </div>
              </button>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{selected ? `Edit ${selected.name}` : 'Add Integration'}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Field label="Name" value={form.name} onChange={(value) => setForm({ ...form, name: value })} placeholder="Prometheus Production" />
            <label className="block text-xs font-medium text-slate-400">
              Type
              <select
                className="mt-1 w-full rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
                value={form.type}
                onChange={(event) => {
                  const selectedType = typeList.find((tool: IntegrationType) => tool.type === event.target.value);
                  setForm({ ...form, type: event.target.value, category: selectedType?.category ?? form.category });
                }}
              >
                {typeList.map((tool: IntegrationType) => <option key={tool.type} value={tool.type}>{tool.type} — {tool.category}</option>)}
              </select>
            </label>
            <Field label="Category" value={form.category ?? ''} onChange={(value) => setForm({ ...form, category: value })} placeholder="Observability" />
            <Field label="Endpoint URL" value={form.endpointUrl} onChange={(value) => setForm({ ...form, endpointUrl: value })} placeholder="https://integration.internal" />
            <Field label="Access Token" value={form.accessToken ?? ''} onChange={(value) => setForm({ ...form, accessToken: value })} placeholder={selected?.hasAccessToken ? 'leave empty to keep encrypted token' : 'optional'} type="password" />
            <div className="grid gap-3 md:grid-cols-2">
              <Field label="Username" value={form.username ?? ''} onChange={(value) => setForm({ ...form, username: value })} placeholder="optional" />
              <Field label="Password" value={form.password ?? ''} onChange={(value) => setForm({ ...form, password: value })} placeholder={selected?.hasPassword ? 'leave empty to keep encrypted password' : 'optional'} type="password" />
            </div>
            <div className="grid gap-3 md:grid-cols-2">
              <Field label="Namespace" value={form.namespace ?? ''} onChange={(value) => setForm({ ...form, namespace: value })} placeholder="monitoring" />
              <Field label="Sync Interval Seconds" value={String(form.syncIntervalSeconds)} onChange={(value) => setForm({ ...form, syncIntervalSeconds: Number(value) })} type="number" />
            </div>
            <div className="flex items-center justify-between rounded-xl border border-slate-800 bg-slate-900/70 px-3 py-2 text-sm text-slate-300">
              <label className="flex items-center gap-2">
                <input type="checkbox" checked={form.tlsVerify} onChange={(event) => setForm({ ...form, tlsVerify: event.target.checked })} />
                TLS verify
              </label>
              <label className="flex items-center gap-2">
                <input type="checkbox" checked={form.enabled} onChange={(event) => setForm({ ...form, enabled: event.target.checked })} />
                Enabled
              </label>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button disabled={isBusy} onClick={submit}>{selected ? 'Save Changes' : 'Add Integration'}</Button>
              <Button disabled={isBusy} className="border-slate-700 bg-slate-900 text-slate-200" onClick={resetForm}>New</Button>
              {selected ? <Button disabled={isBusy} className="border-rose-400/30 bg-rose-400/10 text-rose-100" onClick={() => deleteMutation.mutate(selected.id)}>Delete</Button> : null}
            </div>
            {feedback ? <div className="rounded-xl border border-cyan-400/20 bg-cyan-400/10 p-3 text-xs text-cyan-100">{feedback}</div> : null}
          </CardContent>
        </Card>
      </section>

      {selected ? (
        <section className="grid gap-4 xl:grid-cols-[0.8fr_1.2fr]">
          <Card>
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4">
                <div className="mb-2 font-medium text-slate-100">{selected.name}</div>
                <div className="text-xs leading-5 text-slate-500">{selected.health || 'No health check has been executed.'}</div>
              </div>
              <div className="grid gap-2 md:grid-cols-2">
                <Button disabled={isBusy} onClick={() => enableMutation.mutate({ id: selected.id, enabled: !selected.enabled })}>{selected.enabled ? 'Disable' : 'Enable'}</Button>
                <Button disabled={isBusy || !selected.enabled} onClick={() => testMutation.mutate(selected.id)}>Test Connection</Button>
                <Button disabled={isBusy || !selected.enabled} onClick={() => syncMutation.mutate(selected.id)}>Sync Data</Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Integration Logs</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {(history.data ?? []).slice(0, 3).map((item: ConnectionHistory) => (
                <div key={item.id} className="rounded-xl border border-slate-800 bg-slate-950/70 p-3">
                  <div className="mb-1 flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 text-sm font-medium text-slate-200"><StatusDot status={item.status} />{item.action}</div>
                    <Badge variant={badgeVariant(item.status)}>{item.latencyMs} ms</Badge>
                  </div>
                  <div className="text-xs text-slate-500">{item.message}</div>
                  <div className="mt-1 text-[11px] text-slate-600">{new Date(item.checkedAt).toLocaleString()}</div>
                </div>
              ))}
              {(logs.data ?? []).map((log: IntegrationLog) => (
                <div key={log.id} className="rounded-xl border border-slate-800 bg-slate-900/70 p-3">
                  <div className="mb-1 flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2 text-sm font-medium text-slate-200"><StatusDot status={log.status} />{log.action}</div>
                    <Badge variant={badgeVariant(log.status)}>{log.status}</Badge>
                  </div>
                  <div className="text-xs text-slate-500">{log.message}</div>
                  <div className="mt-1 text-[11px] text-slate-600">{new Date(log.createdAt).toLocaleString()}</div>
                </div>
              ))}
              {!logs.data?.length && !history.data?.length ? <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4 text-sm text-slate-500">No logs yet. Run a test connection or sync.</div> : null}
            </CardContent>
          </Card>
        </section>
      ) : null}
    </div>
  );
}

function Field({ label, value, onChange, placeholder, type = 'text' }: { label: string; value: string; onChange: (value: string) => void; placeholder?: string; type?: string }) {
  return (
    <label className="block text-xs font-medium text-slate-400">
      {label}
      <input
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        className="mt-1 w-full rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50"
      />
    </label>
  );
}
