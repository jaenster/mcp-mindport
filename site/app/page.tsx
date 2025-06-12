'use client';

import { useState, useEffect } from 'react';

interface Stats {
  totalResources: number;
  totalPrompts: number;
  totalDomains: number;
  recentResources: Array<{
    id: string;
    name: string;
    description?: string;
    domain: string;
    updatedAt: string;
  }>;
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/stats')
      .then(res => res.json())
      .then(data => {
        setStats(data.stats);
        setLoading(false);
      })
      .catch(err => {
        console.error('Error loading stats:', err);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <div>Loading dashboard...</div>;
  }

  if (!stats) {
    return <div>Error loading dashboard data</div>;
  }

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Dashboard</h1>
      
      {/* Stats Cards */}
      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
        gap: '1rem',
        marginBottom: '2rem'
      }}>
        <div style={{
          backgroundColor: 'white',
          padding: '1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          textAlign: 'center'
        }}>
          <h3 style={{ margin: '0 0 0.5rem 0', color: '#666' }}>Resources</h3>
          <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333' }}>
            {stats.totalResources}
          </div>
        </div>
        
        <div style={{
          backgroundColor: 'white',
          padding: '1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          textAlign: 'center'
        }}>
          <h3 style={{ margin: '0 0 0.5rem 0', color: '#666' }}>Prompts</h3>
          <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333' }}>
            {stats.totalPrompts}
          </div>
        </div>
        
        <div style={{
          backgroundColor: 'white',
          padding: '1.5rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          textAlign: 'center'
        }}>
          <h3 style={{ margin: '0 0 0.5rem 0', color: '#666' }}>Domains</h3>
          <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#333' }}>
            {stats.totalDomains}
          </div>
        </div>
      </div>

      {/* Recent Resources */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ marginTop: 0 }}>Recent Resources</h2>
        {stats.recentResources.length === 0 ? (
          <p style={{ color: '#666' }}>No resources found. Start by storing some content with the MCP server!</p>
        ) : (
          <div>
            {stats.recentResources.map(resource => (
              <div key={resource.id} style={{
                borderBottom: '1px solid #eee',
                paddingBottom: '1rem',
                marginBottom: '1rem'
              }}>
                <h3 style={{ margin: '0 0 0.5rem 0' }}>
                  <a href={`/resources/${resource.id}`} style={{
                    color: '#0066cc',
                    textDecoration: 'none'
                  }}>
                    {resource.name}
                  </a>
                </h3>
                {resource.description && (
                  <p style={{ margin: '0 0 0.5rem 0', color: '#666' }}>
                    {resource.description}
                  </p>
                )}
                <div style={{ fontSize: '0.875rem', color: '#999' }}>
                  Domain: {resource.domain} • 
                  Updated: {new Date(resource.updatedAt).toLocaleDateString()}
                </div>
              </div>
            ))}
          </div>
        )}
        <div style={{ marginTop: '1rem' }}>
          <a href="/resources" style={{
            color: '#0066cc',
            textDecoration: 'none',
            fontWeight: 'bold'
          }}>
            View all resources →
          </a>
        </div>
      </div>
    </div>
  );
}