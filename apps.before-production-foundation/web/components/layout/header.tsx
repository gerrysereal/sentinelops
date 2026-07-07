import { ShieldCheck } from 'lucide-react';
import { Badge } from '@/components/ui/badge';

export function Header() {
  return (
    <header className="sticky top-0 z-20 border-b border-slate-800/80 bg-slate-950/70 px-6 py-4 backdrop-blur-xl">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold tracking-tight text-slate-50">SentinelOps Dashboard</h1>
          <p className="text-sm text-slate-500">One dashboard for DevOps, DevSecOps, GitOps, and Observability.</p>
        </div>
        <Badge variant="success" className="gap-2">
          <ShieldCheck className="h-3.5 w-3.5" /> Healthy
        </Badge>
      </div>
    </header>
  );
}
