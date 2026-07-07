'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const inputClass = 'rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none focus:border-cyan-400/50';

const defaultSettings = {
  environment: 'production',
  role: 'platform-admin',
  autoRefresh: '30',
  apiBaseUrl: 'http://localhost:8080/api/v1',
  keycloakEnabled: false,
  gitopsEnabled: true,
  securityGate: true,
  observabilityEnabled: true
};

type PlatformSettings = typeof defaultSettings;

export default function SettingsPage() {
  const [settings, setSettings] = useState<PlatformSettings>(defaultSettings);
  const [savedAt, setSavedAt] = useState<string>('');

  useEffect(() => {
    const saved = window.localStorage.getItem('sentinelops-settings');
    if (saved) setSettings({ ...defaultSettings, ...JSON.parse(saved) });
    setSavedAt(window.localStorage.getItem('sentinelops-settings-saved-at') ?? 'Not saved yet');
  }, []);

  function update<K extends keyof PlatformSettings>(key: K, value: PlatformSettings[K]) {
    setSettings((current) => ({ ...current, [key]: value }));
  }

  function save() {
    const timestamp = new Date().toLocaleString();
    window.localStorage.setItem('sentinelops-settings', JSON.stringify(settings));
    window.localStorage.setItem('sentinelops-settings-saved-at', timestamp);
    setSavedAt(timestamp);
  }

  function reset() {
    window.localStorage.removeItem('sentinelops-settings');
    window.localStorage.removeItem('sentinelops-settings-saved-at');
    setSettings(defaultSettings);
    setSavedAt('Not saved yet');
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <CardTitle>Platform Settings</CardTitle>
              <p className="mt-1 text-sm text-slate-500">Local runtime preferences for demo mode. Saved in browser localStorage.</p>
            </div>
            <Badge variant="success">editable</Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-2 text-sm text-slate-300">
              Environment
              <select className={inputClass} value={settings.environment} onChange={(e) => update('environment', e.target.value)}>
                <option value="dev">dev</option>
                <option value="staging">staging</option>
                <option value="production">production</option>
              </select>
            </label>
            <label className="space-y-2 text-sm text-slate-300">
              Active Role
              <select className={inputClass} value={settings.role} onChange={(e) => update('role', e.target.value)}>
                <option value="platform-admin">platform-admin</option>
                <option value="developer">developer</option>
                <option value="security-engineer">security-engineer</option>
                <option value="viewer">viewer</option>
              </select>
            </label>
            <label className="space-y-2 text-sm text-slate-300">
              Dashboard Auto Refresh / seconds
              <input className={inputClass} value={settings.autoRefresh} onChange={(e) => update('autoRefresh', e.target.value)} />
            </label>
            <label className="space-y-2 text-sm text-slate-300">
              API Base URL
              <input className={inputClass} value={settings.apiBaseUrl} onChange={(e) => update('apiBaseUrl', e.target.value)} />
            </label>
          </div>

          <div className="mt-5 grid gap-3 md:grid-cols-2">
            {[
              ['keycloakEnabled', 'Enable Keycloak / OIDC boundary'],
              ['gitopsEnabled', 'Enable GitOps deployment actions'],
              ['securityGate', 'Enable security gate before deployment'],
              ['observabilityEnabled', 'Enable observability widgets']
            ].map(([key, label]) => (
              <label key={key} className="flex items-center justify-between rounded-xl border border-slate-800 bg-slate-900/70 p-4 text-sm text-slate-300">
                {label}
                <input
                  type="checkbox"
                  checked={Boolean(settings[key as keyof PlatformSettings])}
                  onChange={(e) => update(key as keyof PlatformSettings, e.target.checked as never)}
                  className="h-4 w-4 accent-cyan-400"
                />
              </label>
            ))}
          </div>

          <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-slate-800 pt-4">
            <p className="text-xs text-slate-500">Last saved: {savedAt}</p>
            <div className="flex gap-2">
              <Button type="button" className="border-slate-700 bg-slate-900 text-slate-200 hover:bg-slate-800" onClick={reset}>Reset</Button>
              <Button type="button" onClick={save}>Save Settings</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Runtime Preview</CardTitle></CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-4">
          <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Environment</p><p className="text-lg text-slate-100">{settings.environment}</p></div>
          <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Role</p><p className="text-lg text-slate-100">{settings.role}</p></div>
          <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">Auto refresh</p><p className="text-lg text-slate-100">{settings.autoRefresh}s</p></div>
          <div className="rounded-xl border border-slate-800 bg-slate-900/70 p-4"><p className="text-xs text-slate-500">GitOps</p><p className="text-lg text-slate-100">{settings.gitopsEnabled ? 'enabled' : 'disabled'}</p></div>
        </CardContent>
      </Card>
    </div>
  );
}
