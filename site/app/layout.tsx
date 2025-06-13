import type { Metadata } from 'next';
import './globals.css';
import { ThemeProvider } from './components/theme-provider';
import { Navigation } from './components/navigation';

export const metadata: Metadata = {
  title: 'MindPort Browser',
  description: 'Browse and search MindPort MCP data',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <ThemeProvider>
          <Navigation />
          <main className="min-h-screen bg-background">
            <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
              {children}
            </div>
          </main>
        </ThemeProvider>
      </body>
    </html>
  );
}