export function LoadingState({ label = 'Loading platform data...' }: { label?: string }) {
  return <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-8 text-sm text-slate-400">{label}</div>;
}
