import type { Integration } from '@/lib/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

function variant(status: string): 'success' | 'warning' | 'danger' | 'muted' {
  if (status === 'healthy') return 'success';
  if (status === 'warning' || status === 'degraded') return 'warning';
  if (status === 'failed') return 'danger';
  return 'muted';
}

export function IntegrationsCard({ integrations }: { integrations: Integration[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Platform Integrations</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {integrations.map((item) => (
          <div key={item.name} className="rounded-xl border border-slate-800 bg-slate-900/70 p-4">
            <div className="mb-2 flex items-center justify-between gap-2">
              <div className="font-medium text-slate-100">{item.name}</div>
              <Badge variant={variant(item.status)}>{item.status}</Badge>
            </div>
            <div className="text-xs text-slate-500">{item.category}</div>
            <div className="mt-2 truncate text-xs text-slate-400">{item.endpoint}</div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
