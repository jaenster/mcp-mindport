import type { Metadata } from 'next';

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
      <body style={{ 
        fontFamily: 'system-ui, -apple-system, sans-serif',
        margin: 0,
        padding: 0,
        backgroundColor: '#f5f5f5'
      }}>
        <nav style={{
          backgroundColor: '#333',
          color: 'white',
          padding: '1rem',
          marginBottom: '2rem'
        }}>
          <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
            <h1 style={{ margin: 0, fontSize: '1.5rem' }}>
              <a href="/" style={{ color: 'white', textDecoration: 'none' }}>
                🧠 MindPort Browser
              </a>
            </h1>
            <nav style={{ marginTop: '0.5rem' }}>
              <a href="/" style={{ color: '#ccc', marginRight: '1rem', textDecoration: 'none' }}>
                Dashboard
              </a>
              <a href="/resources" style={{ color: '#ccc', marginRight: '1rem', textDecoration: 'none' }}>
                Resources
              </a>
              <a href="/prompts" style={{ color: '#ccc', marginRight: '1rem', textDecoration: 'none' }}>
                Prompts
              </a>
              <a href="/domains" style={{ color: '#ccc', textDecoration: 'none' }}>
                Domains
              </a>
            </nav>
          </div>
        </nav>
        <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 1rem' }}>
          {children}
        </div>
      </body>
    </html>
  );
}