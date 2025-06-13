'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';

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
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Domains</h1>
        <p className="text-muted-foreground">
          Organize your resources into separate contexts or projects
        </p>
      </div>
      
      <Card>
        <CardContent className="pt-6">
          <p className="text-muted-foreground">
            Domains organize your resources into separate contexts or projects. 
            Each domain maintains its own collection of resources and prompts.
          </p>
        </CardContent>
      </Card>

      {domains.length === 0 ? (
        <Card>
          <CardContent className="text-center py-8">
            <svg className="mx-auto h-12 w-12 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <h3 className="mt-2 text-sm font-semibold">No domains found</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Create your first domain using the MCP server!
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {domains.map(domain => (
            <Card key={domain.name} className="hover:shadow-lg transition-shadow">
              <CardHeader>
                <CardTitle className="text-xl">
                  {domain.name}
                </CardTitle>
                {domain.description && (
                  <CardDescription>
                    {domain.description}
                  </CardDescription>
                )}
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between items-center p-3 bg-muted rounded-lg">
                  <div className="text-sm font-medium">
                    <strong>{domain.resourceCount}</strong> resources
                  </div>
                  <div className="text-sm text-muted-foreground">
                    Created {new Date(domain.createdAt).toLocaleDateString()}
                  </div>
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  <Link 
                    href={`/resources?domain=${encodeURIComponent(domain.name)}`}
                    className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-3"
                  >
                    Resources
                  </Link>
                  <Link 
                    href={`/prompts?domain=${encodeURIComponent(domain.name)}`}
                    className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-3"
                  >
                    Prompts
                  </Link>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}