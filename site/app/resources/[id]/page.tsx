'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../components/ui/card';
import { Markdown, detectContentType } from '../../components/ui/markdown';

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

  if (!resource) {
    return (
      <Card>
        <CardContent className="text-center py-8">
          <p>Resource not found</p>
        </CardContent>
      </Card>
    );
  }

  const contentType = detectContentType(resource.content, resource.mimeType);

  return (
    <div className="space-y-6">
      {/* Back Button */}
      <div>
        <Link href="/resources" className="text-sm text-primary hover:underline">
          ← Back to Resources
        </Link>
      </div>

      {/* Resource Header */}
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">{resource.name}</CardTitle>
          {resource.description && (
            <CardDescription className="text-lg">
              {resource.description}
            </CardDescription>
          )}
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Tags */}
          {resource.tags.length > 0 && (
            <div>
              <h3 className="text-sm font-medium text-muted-foreground mb-2 uppercase tracking-wide">
                Tags
              </h3>
              <div className="flex flex-wrap gap-2">
                {resource.tags.map(tag => (
                  <span key={tag} className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary/10 text-primary">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
          
          {/* Metadata */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4 bg-muted rounded-lg">
            <div>
              <span className="font-medium">Domain:</span> {resource.domain}
            </div>
            {resource.mimeType && (
              <div>
                <span className="font-medium">Type:</span> {resource.mimeType}
              </div>
            )}
            {resource.uri && (
              <div className="lg:col-span-3">
                <span className="font-medium">URI:</span>{' '}
                <a 
                  href={resource.uri} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="text-primary hover:underline ml-1 break-all"
                >
                  {resource.uri}
                </a>
              </div>
            )}
            <div>
              <span className="font-medium">Created:</span> {new Date(resource.createdAt).toLocaleString()}
            </div>
            <div>
              <span className="font-medium">Updated:</span> {new Date(resource.updatedAt).toLocaleString()}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Resource Content */}
      <Card>
        <CardHeader>
          <CardTitle>Content</CardTitle>
          <CardDescription>
            {resource.content.length.toLocaleString()} characters
            {contentType !== 'text' && (
              <span className="ml-2 px-2 py-1 bg-secondary text-secondary-foreground rounded text-xs">
                {contentType}
              </span>
            )}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {contentType === 'markdown' ? (
            <div className="border rounded-lg p-4 bg-background">
              <Markdown content={resource.content} />
            </div>
          ) : (
            <div className="rounded-lg bg-muted p-4 font-mono text-sm whitespace-pre-wrap overflow-auto max-h-96 border">
              {resource.content}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}