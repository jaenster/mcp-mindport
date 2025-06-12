import sqlite3 from 'sqlite3';
import { promisify } from 'util';
import * as path from 'path';
import * as os from 'os';

export interface Resource {
  id: string;
  name: string;
  description?: string;
  content: string;
  tags: string[];
  domain: string;
  mimeType?: string;
  uri?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Domain {
  name: string;
  description?: string;
  resourceCount: number;
  createdAt: Date;
}

export interface PromptTemplate {
  id: string;
  name: string;
  description?: string;
  template: string;
  variables: string[];
  domain: string;
  createdAt: Date;
  updatedAt: Date;
}

export class MindPortDB {
  private db: sqlite3.Database;
  private dbGet: (sql: string, ...params: any[]) => Promise<any>;
  private dbAll: (sql: string, ...params: any[]) => Promise<any[]>;

  constructor(dbPath?: string) {
    const defaultPath = path.join(os.homedir(), '.config', 'mindport', 'data', 'storage.db');
    try {
      this.db = new sqlite3.Database(dbPath || defaultPath);
      this.dbGet = promisify(this.db.get.bind(this.db));
      this.dbAll = promisify(this.db.all.bind(this.db));
    } catch (error) {
      console.error('Database connection error:', error);
      throw new Error('Could not connect to MindPort database. Make sure the MCP server has been run at least once to create the database.');
    }
  }

  async listDomains(): Promise<Domain[]> {
    const rows = await this.dbAll(`
      SELECT 
        d.name,
        d.description,
        d.created_at,
        COUNT(r.id) as resource_count
      FROM domains d
      LEFT JOIN resources r ON d.name = r.domain
      GROUP BY d.name, d.description, d.created_at
      ORDER BY d.name
    `);

    return rows.map(row => ({
      name: row.name,
      description: row.description,
      resourceCount: row.resource_count,
      createdAt: new Date(row.created_at)
    }));
  }

  async listResources(domain?: string, limit?: number, offset?: number, search?: string): Promise<Resource[]> {
    let sql = 'SELECT * FROM resources WHERE 1=1';
    const params: any[] = [];
    
    if (domain) {
      sql += ' AND domain = ?';
      params.push(domain);
    }
    
    if (search) {
      sql += ' AND (name LIKE ? OR description LIKE ? OR content LIKE ?)';
      const searchTerm = `%${search}%`;
      params.push(searchTerm, searchTerm, searchTerm);
    }
    
    sql += ' ORDER BY updated_at DESC';
    
    if (limit) {
      sql += ' LIMIT ?';
      params.push(limit);
      
      if (offset) {
        sql += ' OFFSET ?';
        params.push(offset);
      }
    }

    const rows = await this.dbAll(sql, ...params);
    return rows.map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      content: row.content,
      tags: JSON.parse(row.tags || '[]'),
      domain: row.domain,
      mimeType: row.mime_type,
      uri: row.uri,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    }));
  }

  async getResource(id: string): Promise<Resource | null> {
    const row = await this.dbGet('SELECT * FROM resources WHERE id = ?', id);
    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      description: row.description,
      content: row.content,
      tags: JSON.parse(row.tags || '[]'),
      domain: row.domain,
      mimeType: row.mime_type,
      uri: row.uri,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    };
  }

  async listPrompts(domain?: string): Promise<PromptTemplate[]> {
    let sql = 'SELECT * FROM prompts';
    const params: any[] = [];
    
    if (domain) {
      sql += ' WHERE domain = ?';
      params.push(domain);
    }
    
    sql += ' ORDER BY updated_at DESC';

    const rows = await this.dbAll(sql, ...params);
    return rows.map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      template: row.template,
      variables: JSON.parse(row.variables || '[]'),
      domain: row.domain,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    }));
  }

  async getPrompt(id: string): Promise<PromptTemplate | null> {
    const row = await this.dbGet('SELECT * FROM prompts WHERE id = ?', id);
    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      description: row.description,
      template: row.template,
      variables: JSON.parse(row.variables || '[]'),
      domain: row.domain,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    };
  }

  async getStats(): Promise<{
    totalResources: number;
    totalPrompts: number;
    totalDomains: number;
    recentResources: Resource[];
  }> {
    const [resourceCount, promptCount, domainCount] = await Promise.all([
      this.dbGet('SELECT COUNT(*) as count FROM resources'),
      this.dbGet('SELECT COUNT(*) as count FROM prompts'),
      this.dbGet('SELECT COUNT(*) as count FROM domains')
    ]);

    const recentResources = await this.listResources(undefined, 5);

    return {
      totalResources: resourceCount.count,
      totalPrompts: promptCount.count,
      totalDomains: domainCount.count,
      recentResources
    };
  }

  close(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.db.close((err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }
}