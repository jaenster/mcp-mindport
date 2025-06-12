'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';

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

export default function ResourceDetailPage() {
  const params = useParams();
  const [resource, setResource] = useState<Resource | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (params.id) {
      fetch(`/api/resources/${params.id}`)
        .then(res => {
          if (!res.ok) {
            throw new Error('Resource not found');
          }
          return res.json();
        })
        .then(data => {
          setResource(data.resource);
          setLoading(false);
        })
        .catch(err => {
          setError(err.message);
          setLoading(false);
        });
    }
  }, [params.id]);

  if (loading) {
    return <div>Loading resource...</div>;
  }

  if (error) {
    return (
      <div style={{
        backgroundColor: '#fee',
        padding: '1rem',
        borderRadius: '8px',
        border: '1px solid #fcc',
        color: '#c33'
      }}>
        Error: {error}
      </div>
    );
  }

  if (!resource) {
    return <div>Resource not found</div>;
  }

  return (
    <div>
      {/* Back Button */}
      <div style={{ marginBottom: '1rem' }}>
        <a href="/resources" style={{
          color: '#0066cc',
          textDecoration: 'none',
          fontSize: '0.875rem'
        }}>
          ← Back to Resources
        </a>
      </div>

      {/* Resource Header */}
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem'
      }}>
        <h1 style={{ margin: '0 0 1rem 0' }}>{resource.name}</h1>
        
        {resource.description && (
          <p style={{ 
            margin: '0 0 1.5rem 0', 
            color: '#666', 
            fontSize: '1.125rem',
            lineHeight: '1.5'
          }}>
            {resource.description}
          </p>
        )}
        
        {/* Tags */}
        {resource.tags.length > 0 && (
          <div style={{ marginBottom: '1.5rem' }}>
            <h3 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', color: '#666' }}>
              TAGS
            </h3>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
              {resource.tags.map(tag => (
                <span key={tag} style={{
                  backgroundColor: '#e3f2fd',
                  color: '#1976d2',
                  padding: '0.5rem 0.75rem',
                  borderRadius: '16px',
                  fontSize: '0.875rem',
                  fontWeight: 'bold'
                }}>
                  {tag}
                </span>
              ))}
            </div>
          </div>
        )}
        
        {/* Metadata */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: '1rem',
          backgroundColor: '#f8f9fa',
          padding: '1rem',
          borderRadius: '4px'
        }}>
          <div>
            <strong>Domain:</strong> {resource.domain}
          </div>
          {resource.mimeType && (
            <div>
              <strong>Type:</strong> {resource.mimeType}
            </div>
          )}
          {resource.uri && (
            <div>
              <strong>URI:</strong> 
              <a href={resource.uri} target="_blank" rel="noopener noreferrer" style={{
                color: '#0066cc',
                marginLeft: '0.5rem'
              }}>
                {resource.uri}
              </a>
            </div>
          )}
          <div>
            <strong>Created:</strong> {new Date(resource.createdAt).toLocaleString()}
          </div>
          <div>
            <strong>Updated:</strong> {new Date(resource.updatedAt).toLocaleString()}
          </div>
        </div>
      </div>

      {/* Resource Content */}
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ margin: '0 0 1rem 0' }}>Content</h2>
        
        <div style={{
          backgroundColor: '#f8f9fa',
          padding: '1.5rem',
          borderRadius: '8px',
          fontFamily: 'monospace',
          fontSize: '0.875rem',
          lineHeight: '1.6',
          whiteSpace: 'pre-wrap',
          overflow: 'auto',
          maxHeight: '70vh',
          border: '1px solid #e9ecef'
        }}>
          {resource.content}
        </div>
        
        <div style={{ 
          marginTop: '1rem',
          fontSize: '0.75rem',
          color: '#999',
          textAlign: 'right'
        }}>
          {resource.content.length.toLocaleString()} characters
        </div>
      </div>
    </div>
  );
}