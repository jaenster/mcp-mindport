import { NextRequest, NextResponse } from 'next/server';
import { MindPortDB } from '@/lib/db';

// Force dynamic rendering
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  const db = new MindPortDB();
  
  try {
    const domains = await db.listDomains();
    return NextResponse.json({ domains });
  } catch (error) {
    console.error('Error fetching domains:', error);
    return NextResponse.json(
      { error: 'Failed to fetch domains' },
      { status: 500 }
    );
  } finally {
    await db.close();
  }
}