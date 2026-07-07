import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

const artifacts = [
  { name: 'checkout-service', registry: 'Harbor', image: 'harbor.local/sentinelops/checkout-service:1.4.2', sbom: 'generated', signature: 'cosign-valid', scan: 'passed' },
  { name: 'payments-api', registry: 'Harbor', image: 'harbor.local/sentinelops/payments-api:2.1.0', sbom: 'generated', signature: 'cosign-valid', scan: 'failed' },
  { name: 'inventory-worker', registry: 'Harbor', image: 'harbor.local/sentinelops/inventory-worker:1.8.0-rc2', sbom: 'generated', signature: 'cosign-valid', scan: 'passed' }
];

export default function RegistryPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Image and Artifact Registry</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {artifacts.map((artifact) => (
          <div key={artifact.image} className="rounded-xl border border-slate-800 bg-slate-900/70 p-4">
            <div className="mb-2 flex items-center justify-between gap-4">
              <div className="font-medium text-slate-100">{artifact.name}</div>
              <Badge variant={artifact.scan === 'passed' ? 'success' : 'danger'}>{artifact.scan}</Badge>
            </div>
            <div className="text-xs text-slate-500">{artifact.registry}</div>
            <code className="mt-2 block break-all text-xs text-cyan-200">{artifact.image}</code>
            <div className="mt-3 flex gap-2">
              <Badge variant="muted">SBOM: {artifact.sbom}</Badge>
              <Badge variant="success">Signature: {artifact.signature}</Badge>
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
