'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Activity, Bell, Boxes, Gauge, GitBranch, Layers, Lock, Rocket, Settings, Shield } from 'lucide-react';
import { cn } from '@/lib/utils';

const items = [
  { href: '/', label: 'Dashboard', icon: Gauge },
  { href: '/applications', label: 'Applications', icon: Boxes },
  { href: '/pipelines', label: 'CI/CD', icon: GitBranch },
  { href: '/deployments', label: 'Deployment', icon: Rocket },
  { href: '/security', label: 'Security', icon: Shield },
  { href: '/observability', label: 'Observability', icon: Activity },
  { href: '/registry', label: 'Registry', icon: Layers },
  { href: '/settings', label: 'Settings', icon: Settings }
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden min-h-screen w-72 border-r border-slate-800/80 bg-slate-950/80 p-5 lg:block">
      <div className="mb-8 flex items-center gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-cyan-400/30 bg-cyan-400/10">
          <Lock className="h-6 w-6 text-cyan-300" />
        </div>
        <div>
          <div className="text-lg font-bold tracking-wide text-cyan-100">SentinelOps</div>
          <div className="text-xs text-slate-500">Internal Developer Platform</div>
        </div>
      </div>

      <nav className="space-y-2">
        {items.map((item) => {
          const Icon = item.icon;
          const active = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-slate-400 transition hover:bg-slate-900 hover:text-slate-100',
                active && 'border border-cyan-400/30 bg-cyan-400/10 text-cyan-100'
              )}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="mt-8 rounded-2xl border border-slate-800 bg-slate-900/50 p-4">
        <div className="mb-2 flex items-center gap-2 text-xs font-semibold text-slate-300">
          <Bell className="h-4 w-4 text-amber-300" /> Platform Notice
        </div>
        <p className="text-xs leading-5 text-slate-500">
          Local mode is using seeded platform data. Connect Argo CD, Harbor, Keycloak, and observability tools in production.
        </p>
      </div>
    </aside>
  );
}
