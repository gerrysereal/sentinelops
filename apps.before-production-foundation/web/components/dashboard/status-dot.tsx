import { cn } from '@/lib/utils';

const statusClasses: Record<string, string> = {
  healthy: 'bg-emerald-400',
  success: 'bg-emerald-400',
  synced: 'bg-emerald-400',
  warning: 'bg-amber-400',
  degraded: 'bg-amber-400',
  progressing: 'bg-cyan-400',
  running: 'bg-cyan-400',
  pending: 'bg-slate-400',
  failed: 'bg-rose-400',
  high: 'bg-rose-400',
  medium: 'bg-amber-400',
  low: 'bg-sky-400',
  open: 'bg-rose-400',
  resolved: 'bg-emerald-400',
  triaged: 'bg-amber-400',
  connected: 'bg-emerald-400',
  disabled: 'bg-slate-500',
  disconnected: 'bg-rose-400',
  error: 'bg-rose-400',
  unknown: 'bg-amber-400'
};

export function StatusDot({ status }: { status: string }) {
  return <span className={cn('h-2.5 w-2.5 rounded-full', statusClasses[status.toLowerCase()] ?? 'bg-slate-400')} />;
}
