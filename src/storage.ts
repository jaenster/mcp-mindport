import sqlite3 from 'sqlite3';
import { promisify } from 'util';
import { Resource, Domain, PromptTemplate } from './types.js';

export class SQLiteStorage {
  private db: sqlite3.Database;
  private dbRun: (sql: string, ...params: any[]) => Promise<sqlite3.RunResult>;
  private dbGet: (sql: string, ...params: any[]) => Promise<any>;
  private dbAll: (sql: string, ...params: any[]) => Promise<any[]>;

  constructor(private dbPath: string) {
    this.db = new sqlite3.Database(dbPath);
    this.dbRun = promisify(this.db.run.bind(this.db));
    this.dbGet = promisify(this.db.get.bind(this.db));
    this.dbAll = promisify(this.db.all.bind(this.db));
  }

  async initialize(): Promise<void> {
    await this.dbRun(`
      CREATE TABLE IF NOT EXISTS domains (
        name TEXT PRIMARY KEY,
        description TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )
    `);

    await this.dbRun(`
      CREATE TABLE IF NOT EXISTS resources (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        content TEXT NOT NULL,
        tags TEXT NOT NULL,
        domain TEXT NOT NULL,
        mime_type TEXT,
        uri TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (domain) REFERENCES domains (name)
      )
    `);

    await this.dbRun(`
      CREATE TABLE IF NOT EXISTS prompts (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        template TEXT NOT NULL,
        variables TEXT NOT NULL,
        domain TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (domain) REFERENCES domains (name)
      )
    `);

    await this.dbRun(`
      CREATE INDEX IF NOT EXISTS idx_resources_domain ON resources (domain)
    `);

    await this.dbRun(`
      CREATE INDEX IF NOT EXISTS idx_resources_tags ON resources (tags)
    `);

    await this.dbRun(`
      CREATE INDEX IF NOT EXISTS idx_prompts_domain ON prompts (domain)
    `);

    // Create default domain
    await this.createDomain('default', 'Default domain for resources');
  }

  async createDomain(name: string, description?: string): Promise<void> {
    try {
      await this.dbRun(
        'INSERT OR IGNORE INTO domains (name, description) VALUES (?, ?)',
        name,
        description || null
      );
    } catch (error) {
      // Ignore duplicate domain errors
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

  async storeResource(resource: Omit<Resource, 'createdAt' | 'updatedAt'>): Promise<void> {
    await this.createDomain(resource.domain);
    
    const now = new Date().toISOString();
    await this.dbRun(`
      INSERT OR REPLACE INTO resources 
      (id, name, description, content, tags, domain, mime_type, uri, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, 
        COALESCE((SELECT created_at FROM resources WHERE id = ?), ?),
        ?)
    `,
      resource.id,
      resource.name,
      resource.description || null,
      resource.content,
      JSON.stringify(resource.tags),
      resource.domain,
      resource.mimeType || null,
      resource.uri || null,
      resource.id,
      now,
      now
    );
  }

  async getResource(id: string, domain?: string): Promise<Resource | null> {
    let sql = 'SELECT * FROM resources WHERE id = ?';
    const params: any[] = [id];
    
    if (domain) {
      sql += ' AND domain = ?';
      params.push(domain);
    }

    const row = await this.dbGet(sql, ...params);
    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      description: row.description,
      content: row.content,
      tags: JSON.parse(row.tags),
      domain: row.domain,
      mimeType: row.mime_type,
      uri: row.uri,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    };
  }

  async listResources(domain?: string, limit?: number, offset?: number): Promise<Resource[]> {
    let sql = 'SELECT * FROM resources';
    const params: any[] = [];
    
    if (domain) {
      sql += ' WHERE domain = ?';
      params.push(domain);
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
      tags: JSON.parse(row.tags),
      domain: row.domain,
      mimeType: row.mime_type,
      uri: row.uri,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    }));
  }

  async searchResources(query: string, domain?: string): Promise<Resource[]> {
    let sql = `
      SELECT * FROM resources 
      WHERE (name LIKE ? OR description LIKE ? OR content LIKE ?)
    `;
    const searchTerm = `%${query}%`;
    const params: any[] = [searchTerm, searchTerm, searchTerm];
    
    if (domain) {
      sql += ' AND domain = ?';
      params.push(domain);
    }
    
    sql += ' ORDER BY updated_at DESC LIMIT 50';

    const rows = await this.dbAll(sql, ...params);
    return rows.map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      content: row.content,
      tags: JSON.parse(row.tags),
      domain: row.domain,
      mimeType: row.mime_type,
      uri: row.uri,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    }));
  }

  async storePrompt(prompt: Omit<PromptTemplate, 'createdAt' | 'updatedAt'>): Promise<void> {
    await this.createDomain(prompt.domain);
    
    const now = new Date().toISOString();
    await this.dbRun(`
      INSERT OR REPLACE INTO prompts 
      (id, name, description, template, variables, domain, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, 
        COALESCE((SELECT created_at FROM prompts WHERE id = ?), ?),
        ?)
    `,
      prompt.id,
      prompt.name,
      prompt.description || null,
      prompt.template,
      JSON.stringify(prompt.variables),
      prompt.domain,
      prompt.id,
      now,
      now
    );
  }

  async getPrompt(id: string, domain?: string): Promise<PromptTemplate | null> {
    let sql = 'SELECT * FROM prompts WHERE id = ?';
    const params: any[] = [id];
    
    if (domain) {
      sql += ' AND domain = ?';
      params.push(domain);
    }

    const row = await this.dbGet(sql, ...params);
    if (!row) return null;

    return {
      id: row.id,
      name: row.name,
      description: row.description,
      template: row.template,
      variables: JSON.parse(row.variables),
      domain: row.domain,
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
      variables: JSON.parse(row.variables),
      domain: row.domain,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at)
    }));
  }

  async close(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.db.close((err) => {
        if (err) reject(err);
        else resolve();
      });
    });
  }
}