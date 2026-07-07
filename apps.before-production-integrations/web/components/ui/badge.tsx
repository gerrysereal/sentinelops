import { cn } from '@/lib/utils';

type BadgeProps = React.HTMLAttributes<HTMLSpanElement> & {
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'muted';
};

const variants = {
  default: 'border-cyan-400/30 bg-cyan-400/10 text-cyan-200',
  success: 'border-emerald-400/30 bg-emerald-400/10 text-emerald-200',
  warning: 'border-amber-400/30 bg-amber-400/10 text-amber-200',
  danger: 'border-rose-400/30 bg-rose-400/10 text-rose-200',
  muted: 'border-slate-600/40 bg-slate-800/70 text-slate-300'
};

export function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return <span className={cn('inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium', variants[variant], className)} {...props} />;
}
