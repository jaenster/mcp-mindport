'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './components/ui/card';

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

const StatCard = ({ title, value, icon }: { title: string; value: number; icon: React.ReactNode }) => (
  <Card>
    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
      <CardTitle className="text-sm font-medium">{title}</CardTitle>
      {icon}
    </CardHeader>
    <CardContent>
      <div className="text-2xl font-bold">{value}</div>
    </CardContent>
  </Card>
);

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
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-destructive">Error</CardTitle>
            <CardDescription>Failed to load dashboard data</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to MindPort. Here's what's in your knowledge base.
        </p>
      </div>
      
      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <StatCard
          title="Resources"
          value={stats.totalResources}
          icon={
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          }
        />
        <StatCard
          title="Prompts"
          value={stats.totalPrompts}
          icon={
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          }
        />
        <StatCard
          title="Domains"
          value={stats.totalDomains}
          icon={
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          }
        />
      </div>

      {/* Recent Resources */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Resources</CardTitle>
          <CardDescription>
            Your most recently updated resources
          </CardDescription>
        </CardHeader>
        <CardContent>
          {stats.recentResources.length === 0 ? (
            <div className="text-center py-8">
              <svg className="mx-auto h-12 w-12 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <h3 className="mt-2 text-sm font-semibold">No resources</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Start by storing some content with the MCP server!
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              {stats.recentResources.map(resource => (
                <div key={resource.id} className="flex items-start space-x-4 p-4 rounded-lg border">
                  <div className="flex-1 min-w-0">
                    <Link 
                      href={`/resources/${resource.id}`}
                      className="text-sm font-medium text-primary hover:underline"
                    >
                      {resource.name}
                    </Link>
                    {resource.description && (
                      <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
                        {resource.description}
                      </p>
                    )}
                    <div className="mt-2 flex items-center space-x-2 text-xs text-muted-foreground">
                      <span className="inline-flex items-center px-2 py-1 rounded-full bg-secondary text-secondary-foreground">
                        {resource.domain}
                      </span>
                      <span>•</span>
                      <span>Updated {new Date(resource.updatedAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>
              ))}
              <div className="pt-4 border-t">
                <Link
                  href="/resources"
                  className="text-sm font-medium text-primary hover:underline"
                >
                  View all resources →
                </Link>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}