import { NextRequest, NextResponse } from 'next/server';
import { MindPortDB } from '@/lib/db';

// Force dynamic rendering
export const dynamic = 'force-dynamic';

export async function GET(request: NextRequest) {
  const db = new MindPortDB();
  
  try {
    const { searchParams } = new URL(request.url);
    const domain = searchParams.get('domain') || undefined;

    const prompts = await db.listPrompts(domain);
    return NextResponse.json({ prompts });
  } catch (error) {
    console.error('Error fetching prompts:', error);
    return NextResponse.json(
      { error: 'Failed to fetch prompts' },
      { status: 500 }
    );
  } finally {
    await db.close();
  }
}