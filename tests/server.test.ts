import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import { MCPServer } from '../src/server.js';
import { SQLiteStorage } from '../src/storage.js';
import { FuseSearch } from '../src/search.js';
import { Config } from '../src/types.js';

// Mock the MCP SDK to avoid stdio transport issues in tests
vi.mock('@modelcontextprotocol/sdk/server/index.js', () => ({
  Server: vi.fn().mockImplementation(() => ({
    setRequestHandler: vi.fn(),
    connect: vi.fn(),
    close: vi.fn(),
  })),
}));

vi.mock('@modelcontextprotocol/sdk/server/stdio.js', () => ({
  StdioServerTransport: vi.fn(),
}));

describe('MCPServer', () => {
  let server: MCPServer;
  let storage: SQLiteStorage;
  let search: FuseSearch;
  let config: Config;
  let testDbPath: string;

  beforeEach(async () => {
    testDbPath = path.join(os.tmpdir(), `mindport-server-test-${Date.now()}-${Math.random()}.db`);
    
    storage = new SQLiteStorage(testDbPath);
    await storage.initialize();
    
    search = new FuseSearch();
    
    config = {
      server: {
        host: 'localhost',
        port: 8080,
      },
      storage: {
        path: testDbPath,
      },
      search: {
        indexPath: path.join(os.tmpdir(), 'test-search'),
      },
      domain: {
        defaultDomain: 'default',
        isolationMode: 'hierarchical',
        allowCrossDomain: true,
        currentDomain: 'default',
      },
    };

    server = new MCPServer(storage, search, config);
  });

  afterEach(async () => {
    await storage.close();
    try {
      await fs.unlink(testDbPath);
    } catch (error) {
      // Ignore cleanup errors
    }
  });

  describe('Resource Operations', () => {
    it('should store a resource', async () => {
      const result = await server['handleStoreResource']({
        id: 'test-resource',
        name: 'Test Resource',
        description: 'A test resource',
        content: 'This is test content',
        tags: ['test', 'sample'],
        mimeType: 'text/plain',
      });

      expect(result.content[0].text).toContain('stored successfully');
      
      // Verify resource was stored
      const storedResource = await storage.getResource('test-resource');
      expect(storedResource).toBeDefined();
      expect(storedResource?.name).toBe('Test Resource');
    });

    it('should get a resource by ID', async () => {
      // Store a resource first
      await storage.storeResource({
        id: 'get-test',
        name: 'Get Test Resource',
        description: 'Test description',
        content: 'Test content for retrieval',
        tags: ['test'],
        domain: 'default',
      });

      const result = await server['handleGetResource']({
        id: 'get-test',
      });

      expect(result.content[0].text).toContain('Get Test Resource');
      expect(result.content[0].text).toContain('Test content for retrieval');
    });

    it('should handle getting non-existent resource', async () => {
      const result = await server['handleGetResource']({
        id: 'non-existent',
      });

      expect(result.content[0].text).toContain('not found');
    });

    it('should list resources', async () => {
      // Store some resources
      await storage.storeResource({
        id: 'list-test-1',
        name: 'First Resource',
        content: 'Content 1',
        tags: [],
        domain: 'default',
      });
      
      await storage.storeResource({
        id: 'list-test-2',
        name: 'Second Resource',
        content: 'Content 2',
        tags: [],
        domain: 'default',
      });

      const result = await server['handleListResources']({});

      expect(result.content[0].text).toContain('First Resource');
      expect(result.content[0].text).toContain('Second Resource');
    });

    it('should handle empty resource list', async () => {
      const result = await server['handleListResources']({});

      expect(result.content[0].text).toContain('No resources found');
    });

    it('should respect limit and offset in list', async () => {
      // Store multiple resources
      for (let i = 0; i < 5; i++) {
        await storage.storeResource({
          id: `limit-test-${i}`,
          name: `Resource ${i}`,
          content: `Content ${i}`,
          tags: [],
          domain: 'default',
        });
      }

      const result = await server['handleListResources']({
        limit: 2,
        offset: 1,
      });

      // Should contain exactly 2 resources (limited)
      const resourceMatches = result.content[0].text.match(/Resource \d+/g);
      expect(resourceMatches).toHaveLength(2);
    });
  });

  describe('Search Operations', () => {
    beforeEach(async () => {
      // Setup test data for search
      await storage.storeResource({
        id: 'search-1',
        name: 'JavaScript Tutorial',
        description: 'Learn JavaScript basics',
        content: 'JavaScript is a programming language for web development',
        tags: ['javascript', 'tutorial'],
        domain: 'default',
      });
      
      await storage.storeResource({
        id: 'search-2',
        name: 'Python Guide',
        description: 'Python programming guide',
        content: 'Python is a versatile programming language',
        tags: ['python', 'guide'],
        domain: 'default',
      });
    });

    it('should search resources', async () => {
      const result = await server['handleSearchResources']({
        query: 'JavaScript',
        limit: 10,
      });

      expect(result.content[0].text).toContain('JavaScript Tutorial');
      expect(result.content[0].text).toContain('score:');
    });

    it('should handle search with no results', async () => {
      const result = await server['handleSearchResources']({
        query: 'nonexistentquery12345',
        limit: 10,
      });

      expect(result.content[0].text).toContain('No results found');
    });

    it('should perform advanced search with tags', async () => {
      const result = await server['handleAdvancedSearch']({
        query: 'programming',
        tags: ['javascript'],
        exactTags: true,
      });

      expect(result.content[0].text).toContain('JavaScript Tutorial');
      expect(result.content[0].text).not.toContain('Python Guide');
    });

    it('should perform grep search', async () => {
      const result = await server['handleGrep']({
        pattern: 'programming language',
      });

      expect(result.content[0].text).toContain('Pattern matches');
      expect(result.content[0].text).toMatch(/(JavaScript Tutorial|Python Guide)/);
    });

    it('should handle grep with no matches', async () => {
      const result = await server['handleGrep']({
        pattern: 'unmatchablepattern12345',
      });

      expect(result.content[0].text).toContain('No matches found');
    });

    it('should find resources by name pattern', async () => {
      const result = await server['handleFind']({
        pattern: 'Tutorial',
      });

      expect(result.content[0].text).toContain('JavaScript Tutorial');
      expect(result.content[0].text).not.toContain('Python Guide');
    });

    it('should handle find with no matches', async () => {
      const result = await server['handleFind']({
        pattern: 'NonExistentPattern',
      });

      expect(result.content[0].text).toContain('No resources found');
    });
  });

  describe('Domain Operations', () => {
    it('should list domains', async () => {
      await storage.createDomain('test-domain', 'Test domain');
      
      const result = await server['handleListDomains']({});

      expect(result.content[0].text).toContain('default');
      expect(result.content[0].text).toContain('test-domain');
      expect(result.content[0].text).toContain('Current domain: default');
    });

    it('should create a new domain', async () => {
      const result = await server['handleCreateDomain']({
        name: 'new-domain',
        description: 'A new test domain',
      });

      expect(result.content[0].text).toContain('created successfully');
      
      // Verify domain was created
      const domains = await storage.listDomains();
      const newDomain = domains.find(d => d.name === 'new-domain');
      expect(newDomain).toBeDefined();
    });

    it('should switch domains', async () => {
      await storage.createDomain('switch-target', 'Target domain');
      
      const result = await server['handleSwitchDomain']({
        domain: 'switch-target',
      });

      expect(result.content[0].text).toContain('Switched to domain');
      expect(config.domain.currentDomain).toBe('switch-target');
    });

    it('should handle switching to non-existent domain', async () => {
      const result = await server['handleSwitchDomain']({
        domain: 'non-existent-domain',
      });

      expect(result.content[0].text).toContain('does not exist');
      expect(config.domain.currentDomain).toBe('default'); // Should remain unchanged
    });

    it('should provide domain statistics', async () => {
      // Add some resources and prompts to default domain
      await storage.storeResource({
        id: 'stats-resource',
        name: 'Stats Resource',
        content: 'Test content',
        tags: ['tag1', 'tag2'],
        domain: 'default',
      });
      
      await storage.storePrompt({
        id: 'stats-prompt',
        name: 'Stats Prompt',
        template: 'Test template',
        variables: [],
        domain: 'default',
      });

      const result = await server['handleDomainStats']({});

      expect(result.content[0].text).toContain('**Domain:** default');
      expect(result.content[0].text).toContain('**Resources:** 1');
      expect(result.content[0].text).toContain('**Prompts:** 1');
      expect(result.content[0].text).toContain('Top Tags');
    });
  });

  describe('Prompt Operations', () => {
    const samplePrompt = {
      id: 'test-prompt',
      name: 'Test Prompt',
      description: 'A test prompt template',
      template: 'Hello {{name}}, welcome to {{app}}!',
      variables: ['name', 'app'],
    };

    it('should store a prompt', async () => {
      const result = await server['handleStorePrompt'](samplePrompt);

      expect(result.content[0].text).toContain('stored successfully');
      
      // Verify prompt was stored
      const storedPrompt = await storage.getPrompt('test-prompt');
      expect(storedPrompt).toBeDefined();
      expect(storedPrompt?.name).toBe('Test Prompt');
    });

    it('should list prompts', async () => {
      await storage.storePrompt({
        ...samplePrompt,
        domain: 'default',
      });

      const result = await server['handleListPrompts']({});

      expect(result.content[0].text).toContain('Test Prompt');
      expect(result.content[0].text).toContain('Variables: name, app');
    });

    it('should handle empty prompt list', async () => {
      const result = await server['handleListPrompts']({});

      expect(result.content[0].text).toContain('No prompts found');
    });

    it('should get and render a prompt', async () => {
      await storage.storePrompt({
        ...samplePrompt,
        domain: 'default',
      });

      const result = await server['handleGetPrompt']({
        id: 'test-prompt',
        variables: {
          name: 'John',
          app: 'MindPort',
        },
      });

      expect(result.content[0].text).toContain('Test Prompt');
      expect(result.content[0].text).toContain('Hello John, welcome to MindPort!');
    });

    it('should get prompt without variable substitution', async () => {
      await storage.storePrompt({
        ...samplePrompt,
        domain: 'default',
      });

      const result = await server['handleGetPrompt']({
        id: 'test-prompt',
      });

      expect(result.content[0].text).toContain('Hello {{name}}, welcome to {{app}}!');
    });

    it('should handle getting non-existent prompt', async () => {
      const result = await server['handleGetPrompt']({
        id: 'non-existent-prompt',
      });

      expect(result.content[0].text).toContain('not found');
    });
  });

  describe('Tool Configuration', () => {
    it('should provide correct tool definitions', () => {
      const tools = server['getTools']();

      expect(tools).toHaveLength(14);
      
      const toolNames = tools.map(t => t.name);
      expect(toolNames).toContain('store_resource');
      expect(toolNames).toContain('search_resources');
      expect(toolNames).toContain('grep');
      expect(toolNames).toContain('list_domains');
      expect(toolNames).toContain('store_prompt');

      // Check that each tool has required properties
      tools.forEach(tool => {
        expect(tool.name).toBeDefined();
        expect(tool.description).toBeDefined();
        expect(tool.inputSchema).toBeDefined();
        expect(tool.inputSchema.type).toBe('object');
      });
    });
  });
});