import type { SecurityAlert } from '@/lib/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { StatusDot } from './status-dot';

function variantForSeverity(severity: string): 'success' | 'warning' | 'danger' | 'muted' {
  if (severity === 'high') return 'danger';
  if (severity === 'medium') return 'warning';
  if (severity === 'low') return 'muted';
  return 'muted';
}

export function AlertsCard({ alerts }: { alerts: SecurityAlert[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Recent Security Alerts</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {alerts.map((alert) => (
          <div key={alert.id} className="rounded-xl border border-slate-800 bg-slate-900/70 p-3">
            <div className="mb-2 flex items-start justify-between gap-3">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-100">
                <StatusDot status={alert.severity} />
                {alert.title}
              </div>
              <Badge variant={variantForSeverity(alert.severity)}>{alert.severity}</Badge>
            </div>
            <div className="text-xs text-slate-500">{alert.source} • {alert.application} • {alert.status}</div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
