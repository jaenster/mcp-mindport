'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../components/ui/card';

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
  const [copySuccess, setCopySuccess] = useState(false);

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

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(renderedTemplate);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-destructive">Error</CardTitle>
        </CardHeader>
        <CardContent>
          <p>{error}</p>
        </CardContent>
      </Card>
    );
  }

  if (!prompt) {
    return (
      <Card>
        <CardContent className="text-center py-8">
          <p>Prompt not found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back Button */}
      <div>
        <Link href="/prompts" className="text-sm text-primary hover:underline">
          ← Back to Prompts
        </Link>
      </div>

      {/* Prompt Header */}
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">{prompt.name}</CardTitle>
          {prompt.description && (
            <CardDescription className="text-lg">
              {prompt.description}
            </CardDescription>
          )}
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 p-4 bg-muted rounded-lg">
            <div>
              <span className="font-medium">Domain:</span> {prompt.domain}
            </div>
            <div>
              <span className="font-medium">Variables:</span> {prompt.variables.length}
            </div>
            <div>
              <span className="font-medium">Created:</span> {new Date(prompt.createdAt).toLocaleDateString()}
            </div>
            <div>
              <span className="font-medium">Updated:</span> {new Date(prompt.updatedAt).toLocaleDateString()}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Variable Inputs */}
      {prompt.variables.length > 0 && (
        <Card>
          <CardHeader>
            <div className="flex justify-between items-center">
              <div>
                <CardTitle>Template Variables</CardTitle>
                <CardDescription>
                  Enter values for the template variables to see the rendered output
                </CardDescription>
              </div>
              <button
                onClick={resetVariables}
                className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3"
              >
                Reset All
              </button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {prompt.variables.map(variable => (
                <div key={variable} className="space-y-2">
                  <label className="text-sm font-medium font-mono bg-yellow-100 dark:bg-yellow-900 px-2 py-1 rounded">
                    {`{{${variable}}}`}
                  </label>
                  <input
                    type="text"
                    value={variableValues[variable] || ''}
                    onChange={(e) => handleVariableChange(variable, e.target.value)}
                    placeholder={`Enter value for ${variable}...`}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Rendered Template */}
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div>
              <CardTitle>
                {prompt.variables.length > 0 ? 'Rendered Template' : 'Template'}
              </CardTitle>
              <CardDescription>
                {renderedTemplate.length.toLocaleString()} characters
              </CardDescription>
            </div>
            <button
              onClick={copyToClipboard}
              className={`inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 h-9 px-3 ${
                copySuccess 
                  ? 'bg-green-100 text-green-800 border border-green-200' 
                  : 'bg-primary text-primary-foreground hover:bg-primary/90'
              }`}
            >
              {copySuccess ? (
                <>
                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Copied!
                </>
              ) : (
                <>
                  <svg className="h-4 w-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  Copy
                </>
              )}
            </button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg bg-muted p-4 font-mono text-sm whitespace-pre-wrap overflow-auto max-h-96 border">
            {renderedTemplate}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}