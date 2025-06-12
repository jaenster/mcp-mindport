import { NextRequest, NextResponse } from 'next/server';
import { MindPortDB } from '@/lib/db';

// Force dynamic rendering
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  const db = new MindPortDB();
  
  try {
    const stats = await db.getStats();
    return NextResponse.json({ stats });
  } catch (error) {
    console.error('Error fetching stats:', error);
    return NextResponse.json(
      { error: 'Failed to fetch stats' },
      { status: 500 }
    );
  } finally {
    await db.close();
  }
}