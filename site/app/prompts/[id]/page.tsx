'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';

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

export default function PromptDetailPage() {
  const params = useParams();
  const [prompt, setPrompt] = useState<PromptTemplate | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [variableValues, setVariableValues] = useState<Record<string, string>>({});
  const [renderedTemplate, setRenderedTemplate] = useState<string>('');

  useEffect(() => {
    if (params.id) {
      fetch(`/api/prompts/${params.id}`)
        .then(res => {
          if (!res.ok) {
            throw new Error('Prompt not found');
          }
          return res.json();
        })
        .then(data => {
          setPrompt(data.prompt);
          setRenderedTemplate(data.prompt.template);
          setLoading(false);
        })
        .catch(err => {
          setError(err.message);
          setLoading(false);
        });
    }
  }, [params.id]);

  const handleVariableChange = (variable: string, value: string) => {
    const newValues = { ...variableValues, [variable]: value };
    setVariableValues(newValues);
    
    if (prompt) {
      let rendered = prompt.template;
      Object.entries(newValues).forEach(([key, val]) => {
        const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
        rendered = rendered.replace(regex, val);
      });
      setRenderedTemplate(rendered);
    }
  };

  const resetVariables = () => {
    setVariableValues({});
    if (prompt) {
      setRenderedTemplate(prompt.template);
    }
  };

  if (loading) {
    return <div>Loading prompt...</div>;
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

  if (!prompt) {
    return <div>Prompt not found</div>;
  }

  return (
    <div>
      {/* Back Button */}
      <div style={{ marginBottom: '1rem' }}>
        <a href="/prompts" style={{
          color: '#0066cc',
          textDecoration: 'none',
          fontSize: '0.875rem'
        }}>
          ← Back to Prompts
        </a>
      </div>

      {/* Prompt Header */}
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem'
      }}>
        <h1 style={{ margin: '0 0 1rem 0' }}>{prompt.name}</h1>
        
        {prompt.description && (
          <p style={{ 
            margin: '0 0 1.5rem 0', 
            color: '#666', 
            fontSize: '1.125rem',
            lineHeight: '1.5'
          }}>
            {prompt.description}
          </p>
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
            <strong>Domain:</strong> {prompt.domain}
          </div>
          <div>
            <strong>Variables:</strong> {prompt.variables.length}
          </div>
          <div>
            <strong>Created:</strong> {new Date(prompt.createdAt).toLocaleString()}
          </div>
          <div>
            <strong>Updated:</strong> {new Date(prompt.updatedAt).toLocaleString()}
          </div>
        </div>
      </div>

      {/* Variable Inputs */}
      {prompt.variables.length > 0 && (
        <div style={{
          backgroundColor: 'white',
          padding: '2rem',
          borderRadius: '8px',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          marginBottom: '2rem'
        }}>
          <div style={{ 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            marginBottom: '1rem'
          }}>
            <h2 style={{ margin: 0 }}>Template Variables</h2>
            <button
              onClick={resetVariables}
              style={{
                backgroundColor: '#6c757d',
                color: 'white',
                border: 'none',
                padding: '0.5rem 1rem',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '0.875rem'
              }}
            >
              Reset All
            </button>
          </div>
          
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
            gap: '1rem'
          }}>
            {prompt.variables.map(variable => (
              <div key={variable}>
                <label style={{
                  display: 'block',
                  marginBottom: '0.5rem',
                  fontWeight: 'bold',
                  color: '#333',
                  fontFamily: 'monospace'
                }}>
                  {`{{${variable}}}`}
                </label>
                <input
                  type="text"
                  value={variableValues[variable] || ''}
                  onChange={(e) => handleVariableChange(variable, e.target.value)}
                  placeholder={`Enter value for ${variable}...`}
                  style={{
                    width: '100%',
                    padding: '0.5rem',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Rendered Template */}
      <div style={{
        backgroundColor: 'white',
        padding: '2rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ margin: '0 0 1rem 0' }}>
          {prompt.variables.length > 0 ? 'Rendered Template' : 'Template'}
        </h2>
        
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
          {renderedTemplate}
        </div>
        
        <div style={{ 
          marginTop: '1rem',
          fontSize: '0.75rem',
          color: '#999',
          textAlign: 'right'
        }}>
          {renderedTemplate.length.toLocaleString()} characters
        </div>
        
        {/* Copy Button */}
        <div style={{ marginTop: '1rem' }}>
          <button
            onClick={() => {
              navigator.clipboard.writeText(renderedTemplate);
              // Simple feedback - could be improved with toast
              alert('Template copied to clipboard!');
            }}
            style={{
              backgroundColor: '#0066cc',
              color: 'white',
              border: 'none',
              padding: '0.75rem 1.5rem',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: 'bold'
            }}
          >
            📋 Copy to Clipboard
          </button>
        </div>
      </div>
    </div>
  );
}