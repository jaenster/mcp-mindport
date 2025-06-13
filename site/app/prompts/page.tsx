'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';

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
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Prompt Templates</h1>
        <p className="text-muted-foreground">
          Manage and browse your reusable prompt templates
        </p>
      </div>
      
      {/* Filter */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="max-w-xs space-y-2">
            <label className="text-sm font-medium">
              Domain
            </label>
            <select
              value={selectedDomain}
              onChange={(e) => setSelectedDomain(e.target.value)}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="">All domains</option>
              {domains.map(domain => (
                <option key={domain.name} value={domain.name}>
                  {domain.name}
                </option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      {/* Prompts List */}
      {loading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div>
          {prompts.length === 0 ? (
            <Card>
              <CardContent className="text-center py-8">
                <svg className="mx-auto h-12 w-12 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <h3 className="mt-2 text-sm font-semibold">No prompt templates found</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Create prompts using the MCP server to see them here!
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {prompts.map(prompt => (
                <Card key={prompt.id}>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      <Link 
                        href={`/prompts/${prompt.id}`}
                        className="text-primary hover:underline"
                      >
                        {prompt.name}
                      </Link>
                    </CardTitle>
                    {prompt.description && (
                      <CardDescription>
                        {prompt.description}
                      </CardDescription>
                    )}
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="rounded-md bg-muted p-4 font-mono text-sm whitespace-pre-wrap overflow-auto max-h-40">
                      {truncateTemplate(prompt.template)}
                    </div>
                    
                    {/* Variables */}
                    {prompt.variables.length > 0 && (
                      <div>
                        <h4 className="text-sm font-medium text-muted-foreground mb-2">
                          Variables:
                        </h4>
                        <div className="flex flex-wrap gap-2">
                          {prompt.variables.map(variable => (
                            <span key={variable} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-mono font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
                              {`{{${variable}}}`}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                    
                    <div className="flex justify-between items-center text-sm text-muted-foreground pt-2 border-t">
                      <span className="inline-flex items-center px-2 py-1 rounded-full bg-secondary text-secondary-foreground text-xs">
                        {prompt.domain}
                      </span>
                      <div>
                        Updated {new Date(prompt.updatedAt).toLocaleDateString()}
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}