import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatusDot } from './status-dot';

export function StatusList({ title, data }: { title: string; data: Record<string, number> }) {
  const entries = Object.entries(data);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {entries.map(([status, value]) => (
          <div key={status} className="flex items-center justify-between rounded-xl bg-slate-900/70 px-3 py-2">
            <div className="flex items-center gap-2 text-sm capitalize text-slate-300">
              <StatusDot status={status} />
              {status.replaceAll('-', ' ')}
            </div>
            <div className="font-semibold text-slate-100">{value}</div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
