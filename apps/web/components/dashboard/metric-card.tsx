import { Card, CardContent } from '@/components/ui/card';

export function MetricCard({ label, value, hint }: { label: string; value: number | string; hint?: string }) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="text-xs uppercase tracking-[0.2em] text-slate-500">{label}</div>
        <div className="mt-3 text-3xl font-bold text-slate-50">{value}</div>
        {hint ? <div className="mt-2 text-xs text-slate-500">{hint}</div> : null}
      </CardContent>
    </Card>
  );
}
