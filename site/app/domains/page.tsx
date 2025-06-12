'use client';

import { useState, useEffect } from 'react';

interface Domain {
  name: string;
  description?: string;
  resourceCount: number;
  createdAt: string;
}

export default function DomainsPage() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/domains')
      .then(res => res.json())
      .then(data => {
        setDomains(data.domains || []);
        setLoading(false);
      })
      .catch(err => {
        console.error('Error loading domains:', err);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <div>Loading domains...</div>;
  }

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Domains</h1>
      
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem'
      }}>
        <p style={{ margin: 0, color: '#666' }}>
          Domains organize your resources into separate contexts or projects. 
          Each domain maintains its own collection of resources and prompts.
        </p>
      </div>

      {domains.length === 0 ? (
        <div style={{
          backgroundColor: 'white',
          padding: '2rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          textAlign: 'center',
          color: '#666'
        }}>
          No domains found. Create your first domain using the MCP server!
        </div>
      ) : (
        <div style={{ 
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
          gap: '1.5rem'
        }}>
          {domains.map(domain => (
            <div key={domain.name} style={{
              backgroundColor: 'white',
              padding: '1.5rem',
              borderRadius: '8px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              border: '1px solid #e9ecef'
            }}>
              <h3 style={{ 
                margin: '0 0 0.5rem 0',
                color: '#333',
                fontSize: '1.25rem'
              }}>
                {domain.name}
              </h3>
              
              {domain.description && (
                <p style={{ 
                  margin: '0 0 1rem 0', 
                  color: '#666',
                  lineHeight: '1.5'
                }}>
                  {domain.description}
                </p>
              )}
              
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                backgroundColor: '#f8f9fa',
                padding: '0.75rem',
                borderRadius: '4px',
                marginBottom: '1rem'
              }}>
                <div>
                  <strong>{domain.resourceCount}</strong> resources
                </div>
                <div style={{ fontSize: '0.875rem', color: '#666' }}>
                  Created: {new Date(domain.createdAt).toLocaleDateString()}
                </div>
              </div>
              
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <a 
                  href={`/resources?domain=${encodeURIComponent(domain.name)}`}
                  style={{
                    backgroundColor: '#0066cc',
                    color: 'white',
                    padding: '0.5rem 1rem',
                    borderRadius: '4px',
                    textDecoration: 'none',
                    fontSize: '0.875rem',
                    fontWeight: 'bold',
                    textAlign: 'center',
                    flex: 1
                  }}
                >
                  View Resources
                </a>
                <a 
                  href={`/prompts?domain=${encodeURIComponent(domain.name)}`}
                  style={{
                    backgroundColor: '#6c757d',
                    color: 'white',
                    padding: '0.5rem 1rem',
                    borderRadius: '4px',
                    textDecoration: 'none',
                    fontSize: '0.875rem',
                    fontWeight: 'bold',
                    textAlign: 'center',
                    flex: 1
                  }}
                >
                  View Prompts
                </a>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}