import type { Metadata } from 'next';
import './globals.css';
import { Providers } from '@/components/layout/providers';
import { Shell } from '@/components/layout/shell';

export const metadata: Metadata = {
  title: 'SentinelOps',
  description: 'Open-source Internal Developer Platform for DevSecOps and GitOps.'
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className="dark">
      <body>
        <Providers>
          <Shell>{children}</Shell>
        </Providers>
      </body>
    </html>
  );
}
