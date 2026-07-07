export function ErrorState({ message }: { message: string }) {
  return <div className="rounded-2xl border border-rose-500/30 bg-rose-500/10 p-8 text-sm text-rose-100">{message}</div>;
}
