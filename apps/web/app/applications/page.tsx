'use client';

import { FormEvent, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Application } from '@/lib/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Column, DataTable } from '@/components/dashboard/data-table';
import { ErrorState } from '@/components/dashboard/error-state';
import { LoadingState } from '@/components/dashboard/loading-state';

const inputClass = 'rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50';

export default function ApplicationsPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState({ name: '', owner: 'platform-team', repository: '', environment: 'production' });
  const { data, isLoading, error } = useQuery({ queryKey: ['applications'], queryFn: api.applications });

  const createMutation = useMutation({
    mutationFn: api.createApplication,
    onSuccess: async () => {
      setForm({ name: '', owner: 'platform-team', repository: '', environment: 'production' });
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['applications'] }),
        queryClient.invalidateQueries({ queryKey: ['overview'] })
      ]);
    }
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    createMutation.mutate(form);
  }

  if (isLoading) return <LoadingState />;
  if (error || !data) return <ErrorState message="Unable to load applications." />;

  const columns: Column<Application>[] = [
    { key: 'name', header: 'Application', cell: (row) => <span className="font-medium text-slate-100">{row.name}</span> },
    { key: 'owner', header: 'Owner', cell: (row) => row.owner },
    { key: 'env', header: 'Environment', cell: (row) => <Badge variant="muted">{row.environment}</Badge> },
    { key: 'status', header: 'Status', cell: (row) => <Badge variant={row.status === 'healthy' ? 'success' : 'warning'}>{row.status}</Badge> },
    { key: 'repo', header: 'Repository', cell: (row) => <span className="text-slate-500">{row.repository}</span> }
  ];

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Create Application</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={submit} className="grid gap-3 md:grid-cols-5">
            <input className={inputClass} placeholder="service-name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            <input className={inputClass} placeholder="owner" value={form.owner} onChange={(e) => setForm({ ...form, owner: e.target.value })} required />
            <input className={`${inputClass} md:col-span-2`} placeholder="https://github.com/org/repo" value={form.repository} onChange={(e) => setForm({ ...form, repository: e.target.value })} required />
            <select className={inputClass} value={form.environment} onChange={(e) => setForm({ ...form, environment: e.target.value })}>
              <option value="dev">dev</option>
              <option value="staging">staging</option>
              <option value="production">production</option>
            </select>
            <Button disabled={createMutation.isPending}>{createMutation.isPending ? 'Creating...' : 'Register App'}</Button>
          </form>
          {createMutation.error ? <p className="mt-3 text-xs text-rose-300">{createMutation.error.message}</p> : null}
        </CardContent>
      </Card>
      <DataTable columns={columns} data={data} />
    </div>
  );
}
