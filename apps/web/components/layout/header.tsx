'use client';

import { useQuery } from '@tanstack/react-query';
import { ShieldCheck, ShieldAlert } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';

export function Header() {
  const health = useQuery({ queryKey: ['platform-health'], queryFn: api.platformHealth });
  const healthy = health.data?.status === 'ready' || health.data?.status === 'ok';
  const label = health.isLoading ? 'Checking' : healthy ? 'Healthy' : 'Degraded';
  const Icon = healthy ? ShieldCheck : ShieldAlert;

  return (
    <header className="sticky top-0 z-20 border-b border-slate-800/80 bg-slate-950/70 px-6 py-4 backdrop-blur-xl">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight text-slate-50">SentinelOps Dashboard</h1>
          <p className="text-sm text-slate-500">One dashboard for DevOps, DevSecOps, GitOps, and Observability.</p>
        </div>
        <Badge variant={healthy ? 'success' : health.isLoading ? 'muted' : 'warning'} className="gap-2">
          <Icon className="h-3.5 w-3.5" /> {label}
        </Badge>
      </div>
    </header>
  );
}
