import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function ProgressCard({ data }: { data: Record<string, number> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Resource Usage</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {Object.entries(data).map(([key, value]) => (
          <div key={key}>
            <div className="mb-2 flex justify-between text-sm capitalize text-slate-300">
              <span>{key}</span>
              <span>{value}%</span>
            </div>
            <div className="h-2 rounded-full bg-slate-800">
              <div className="h-2 rounded-full bg-cyan-300" style={{ width: `${value}%` }} />
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
