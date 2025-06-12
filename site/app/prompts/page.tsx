'use client';

import { useState, useEffect } from 'react';

interface PromptTemplate {
  id: string;
  name: string;
  description?: string;
  template: string;
  variables: string[];
  domain: string;
  createdAt: string;
  updatedAt: string;
}

interface Domain {
  name: string;
  description?: string;
  resourceCount: number;
}

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<PromptTemplate[]>([]);
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDomain, setSelectedDomain] = useState<string>('');

  const loadPrompts = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (selectedDomain) params.append('domain', selectedDomain);
      
      const res = await fetch(`/api/prompts?${params}`);
      const data = await res.json();
      setPrompts(data.prompts || []);
    } catch (error) {
      console.error('Error loading prompts:', error);
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
    loadPrompts();
  }, [selectedDomain]);

  const truncateTemplate = (template: string, maxLength: number = 200) => {
    if (template.length <= maxLength) return template;
    return template.substring(0, maxLength) + '...';
  };

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Prompt Templates</h1>
      
      {/* Filter */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem'
      }}>
        <div style={{ maxWidth: '300px' }}>
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
                {domain.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Prompts List */}
      {loading ? (
        <div>Loading prompts...</div>
      ) : (
        <div>
          {prompts.length === 0 ? (
            <div style={{
              backgroundColor: 'white',
              padding: '2rem',
              borderRadius: '8px',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              textAlign: 'center',
              color: '#666'
            }}>
              No prompt templates found. Create prompts using the MCP server to see them here!
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {prompts.map(prompt => (
                <div key={prompt.id} style={{
                  backgroundColor: 'white',
                  padding: '1.5rem',
                  borderRadius: '8px',
                  boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
                }}>
                  <h3 style={{ margin: '0 0 0.5rem 0' }}>
                    <a href={`/prompts/${prompt.id}`} style={{
                      color: '#0066cc',
                      textDecoration: 'none'
                    }}>
                      {prompt.name}
                    </a>
                  </h3>
                  
                  {prompt.description && (
                    <p style={{ margin: '0 0 1rem 0', color: '#666' }}>
                      {prompt.description}
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
                    {truncateTemplate(prompt.template)}
                  </div>
                  
                  {/* Variables */}
                  {prompt.variables.length > 0 && (
                    <div style={{ marginBottom: '1rem' }}>
                      <h4 style={{ 
                        margin: '0 0 0.5rem 0', 
                        fontSize: '0.875rem', 
                        color: '#666' 
                      }}>
                        Variables:
                      </h4>
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                        {prompt.variables.map(variable => (
                          <span key={variable} style={{
                            backgroundColor: '#fff3cd',
                            color: '#856404',
                            padding: '0.25rem 0.5rem',
                            borderRadius: '12px',
                            fontSize: '0.75rem',
                            fontWeight: 'bold',
                            fontFamily: 'monospace'
                          }}>
                            {`{{${variable}}}`}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  <div style={{ 
                    fontSize: '0.875rem', 
                    color: '#999',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                  }}>
                    <div>
                      Domain: <strong>{prompt.domain}</strong>
                    </div>
                    <div>
                      Updated: {new Date(prompt.updatedAt).toLocaleDateString()}
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