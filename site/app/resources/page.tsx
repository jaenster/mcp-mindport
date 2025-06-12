'use client';

import { useState, useEffect } from 'react';

interface Resource {
  id: string;
  name: string;
  description?: string;
  content: string;
  tags: string[];
  domain: string;
  mimeType?: string;
  uri?: string;
  createdAt: string;
  updatedAt: string;
}

interface Domain {
  name: string;
  description?: string;
  resourceCount: number;
}

export default function ResourcesPage() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDomain, setSelectedDomain] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const loadResources = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (selectedDomain) params.append('domain', selectedDomain);
      if (searchQuery) params.append('search', searchQuery);
      
      const res = await fetch(`/api/resources?${params}`);
      const data = await res.json();
      setResources(data.resources || []);
    } catch (error) {
      console.error('Error loading resources:', error);
    }
    setLoading(false);
  };

  const loadDomains = async () => {
    try {
      const res = await fetch('/api/domains');
      const data = await res.json();
      setDomains(data.domains || []);
    } catch (error) {
      console.error('Error loading domains:', error);
    }
  };

  useEffect(() => {
    loadDomains();
  }, []);

  useEffect(() => {
    loadResources();
  }, [selectedDomain, searchQuery]);

  const truncateContent = (content: string, maxLength: number = 200) => {
    if (content.length <= maxLength) return content;
    return content.substring(0, maxLength) + '...';
  };

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Resources</h1>
      
      {/* Filters */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem'
      }}>
        <div style={{ 
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: '1rem',
          alignItems: 'end'
        }}>
          <div>
            <label style={{ 
              display: 'block', 
              marginBottom: '0.5rem', 
              fontWeight: 'bold',
              color: '#333'
            }}>
              Search:
            </label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search resources..."
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
                fontSize: '1rem'
              }}
            />
          </div>
          
          <div>
            <label style={{ 
              display: 'block', 
              marginBottom: '0.5rem', 
              fontWeight: 'bold',
              color: '#333'
            }}>
              Domain:
            </label>
            <select
              value={selectedDomain}
              onChange={(e) => setSelectedDomain(e.target.value)}
              style={{
                width: '100%',
                padding: '0.5rem',
                border: '1px solid #ddd',
                borderRadius: '4px',
                fontSize: '1rem',
                backgroundColor: 'white'
              }}
            >
              <option value="">All domains</option>
              {domains.map(domain => (
                <option key={domain.name} value={domain.name}>
                  {domain.name} ({domain.resourceCount})
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Resources List */}
      {loading ? (
        <div>Loading resources...</div>
      ) : (
        <div>
          {resources.length === 0 ? (
            <div style={{
              backgroundColor: 'white',
              padding: '2rem',
              borderRadius: '8px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              textAlign: 'center',
              color: '#666'
            }}>
              No resources found. Try adjusting your search or domain filter.
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {resources.map(resource => (
                <div key={resource.id} style={{
                  backgroundColor: 'white',
                  padding: '1.5rem',
                  borderRadius: '8px',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
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
                    <p style={{ margin: '0 0 1rem 0', color: '#666' }}>
                      {resource.description}
                    </p>
                  )}
                  
                  <div style={{
                    backgroundColor: '#f8f9fa',
                    padding: '1rem',
                    borderRadius: '4px',
                    marginBottom: '1rem',
                    fontFamily: 'monospace',
                    fontSize: '0.875rem',
                    whiteSpace: 'pre-wrap',
                    overflow: 'auto'
                  }}>
                    {truncateContent(resource.content)}
                  </div>
                  
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem', marginBottom: '1rem' }}>
                    {resource.tags.map(tag => (
                      <span key={tag} style={{
                        backgroundColor: '#e3f2fd',
                        color: '#1976d2',
                        padding: '0.25rem 0.5rem',
                        borderRadius: '12px',
                        fontSize: '0.75rem',
                        fontWeight: 'bold'
                      }}>
                        {tag}
                      </span>
                    ))}
                  </div>
                  
                  <div style={{ 
                    fontSize: '0.875rem', 
                    color: '#999',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                  }}>
                    <div>
                      Domain: <strong>{resource.domain}</strong>
                      {resource.mimeType && (
                        <span> • Type: {resource.mimeType}</span>
                      )}
                    </div>
                    <div>
                      Updated: {new Date(resource.updatedAt).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}