import { NextRequest, NextResponse } from 'next/server';
import { MindPortDB } from '@/lib/db';

// Force dynamic rendering
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  const db = new MindPortDB();
  
  try {
    const { searchParams } = new URL(request.url);
    const domain = searchParams.get('domain') || undefined;
    const search = searchParams.get('search') || undefined;
    const limit = searchParams.get('limit') ? parseInt(searchParams.get('limit')!) : 50;
    const offset = searchParams.get('offset') ? parseInt(searchParams.get('offset')!) : 0;

    const resources = await db.listResources(domain, limit, offset, search);
    
    return NextResponse.json({
      resources,
      pagination: {
        limit,
        offset,
        hasMore: resources.length === limit
      }
    });
  } catch (error) {
    console.error('Error fetching resources:', error);
    return NextResponse.json(
      { error: 'Failed to fetch resources' },
      { status: 500 }
    );
  } finally {
    await db.close();
  }
}