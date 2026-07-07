import { Header } from './header';
import { Sidebar } from './sidebar';

export function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid-pattern flex min-h-screen">
      <Sidebar />
      <main className="min-w-0 flex-1">
        <Header />
        <div className="p-6">{children}</div>
      </main>
    </div>
  );
}
