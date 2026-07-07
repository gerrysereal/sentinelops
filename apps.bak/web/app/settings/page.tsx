import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

const settings = [
  ['Authentication', 'Keycloak/OIDC boundary is present. Local development auth is disabled by default.'],
  ['GitOps', 'Argo CD application manifest is available under gitops/argocd.'],
  ['Policy', 'OPA Gatekeeper is represented as an integration and can be installed through cluster addons.'],
  ['Secrets', 'Vault/OpenBao is the intended backend for dynamic secrets and audit logs.'],
  ['Security Scanning', 'CI workflow includes scanner stages for Trivy, Semgrep, and Gitleaks.']
];

export default function SettingsPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Platform Settings</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {settings.map(([name, description]) => (
          <div key={name} className="rounded-xl border border-slate-800 bg-slate-900/70 p-4">
            <div className="mb-2 flex items-center justify-between">
              <div className="font-medium text-slate-100">{name}</div>
              <Badge variant="success">configured</Badge>
            </div>
            <p className="text-sm text-slate-500">{description}</p>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
