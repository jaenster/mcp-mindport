'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Markdown, detectContentType } from '../components/ui/markdown';

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
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Resources</h1>
        <p className="text-muted-foreground">
          Browse and search your stored resources
        </p>
      </div>
      
      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">
                Search
              </label>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search resources..."
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
            
            <div className="space-y-2">
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
                    {domain.name} ({domain.resourceCount})
                  </option>
                ))}
              </select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Resources List */}
      {loading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      ) : (
        <div>
          {resources.length === 0 ? (
            <Card>
              <CardContent className="text-center py-8">
                <svg className="mx-auto h-12 w-12 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <h3 className="mt-2 text-sm font-semibold">No resources found</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Try adjusting your search or domain filter.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              {resources.map(resource => (
                <Card key={resource.id}>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      <Link 
                        href={`/resources/${resource.id}`}
                        className="text-primary hover:underline"
                      >
                        {resource.name}
                      </Link>
                    </CardTitle>
                    {resource.description && (
                      <CardDescription>
                        {resource.description}
                      </CardDescription>
                    )}
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {(() => {
                      const contentType = detectContentType(resource.content, resource.mimeType);
                      const truncatedContent = truncateContent(resource.content);
                      
                      if (contentType === 'markdown') {
                        return (
                          <div className="rounded-md border bg-muted/30 p-4 max-h-40 overflow-auto">
                            <Markdown content={truncatedContent} />
                          </div>
                        );
                      } else {
                        return (
                          <div className="rounded-md bg-muted p-4 font-mono text-sm whitespace-pre-wrap overflow-auto max-h-40">
                            {truncatedContent}
                          </div>
                        );
                      }
                    })()}
                    
                    {resource.tags.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {resource.tags.map(tag => (
                          <span key={tag} className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}
                    
                    <div className="flex justify-between items-center text-sm text-muted-foreground pt-2 border-t">
                      <div className="space-x-2">
                        <span className="inline-flex items-center px-2 py-1 rounded-full bg-secondary text-secondary-foreground text-xs">
                          {resource.domain}
                        </span>
                        {resource.mimeType && (
                          <span>• {resource.mimeType}</span>
                        )}
                      </div>
                      <div>
                        Updated {new Date(resource.updatedAt).toLocaleDateString()}
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