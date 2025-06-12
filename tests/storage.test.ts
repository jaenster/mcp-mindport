import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import { SQLiteStorage } from '../src/storage.js';
import { Resource, PromptTemplate } from '../src/types.js';

describe('SQLiteStorage', () => {
  let storage: SQLiteStorage;
  let testDbPath: string;

  beforeEach(async () => {
    testDbPath = path.join(os.tmpdir(), `mindport-storage-test-${Date.now()}-${Math.random()}.db`);
    storage = new SQLiteStorage(testDbPath);
    await storage.initialize();
  });

  afterEach(async () => {
    await storage.close();
    try {
      await fs.unlink(testDbPath);
    } catch (error) {
      // Ignore cleanup errors
    }
  });

  describe('Domain Management', () => {
    it('should create a default domain on initialization', async () => {
      const domains = await storage.listDomains();
      expect(domains).toHaveLength(1);
      expect(domains[0].name).toBe('default');
      expect(domains[0].resourceCount).toBe(0);
    });

    it('should create new domains', async () => {
      await storage.createDomain('test-domain', 'Test domain description');
      
      const domains = await storage.listDomains();
      const testDomain = domains.find(d => d.name === 'test-domain');
      
      expect(testDomain).toBeDefined();
      expect(testDomain?.description).toBe('Test domain description');
      expect(testDomain?.resourceCount).toBe(0);
    });

    it('should handle duplicate domain creation gracefully', async () => {
      await storage.createDomain('duplicate', 'First description');
      await storage.createDomain('duplicate', 'Second description');
      
      const domains = await storage.listDomains();
      const duplicateDomains = domains.filter(d => d.name === 'duplicate');
      
      expect(duplicateDomains).toHaveLength(1);
    });
  });

  describe('Resource Management', () => {
    const sampleResource: Omit<Resource, 'createdAt' | 'updatedAt'> = {
      id: 'test-resource-1',
      name: 'Test Resource',
      description: 'A test resource for testing',
      content: 'This is the content of the test resource',
      tags: ['test', 'sample', 'resource'],
      domain: 'default',
      mimeType: 'text/plain',
      uri: 'test://resource/1'
    };

    it('should store a resource', async () => {
      await storage.storeResource(sampleResource);
      
      const retrieved = await storage.getResource('test-resource-1');
      
      expect(retrieved).toBeDefined();
      expect(retrieved?.name).toBe(sampleResource.name);
      expect(retrieved?.content).toBe(sampleResource.content);
      expect(retrieved?.tags).toEqual(sampleResource.tags);
      expect(retrieved?.domain).toBe(sampleResource.domain);
      expect(retrieved?.createdAt).toBeInstanceOf(Date);
      expect(retrieved?.updatedAt).toBeInstanceOf(Date);
    });

    it('should update existing resource', async () => {
      await storage.storeResource(sampleResource);
      
      const updatedResource = {
        ...sampleResource,
        name: 'Updated Test Resource',
        content: 'Updated content'
      };
      
      await storage.storeResource(updatedResource);
      const retrieved = await storage.getResource('test-resource-1');
      
      expect(retrieved?.name).toBe('Updated Test Resource');
      expect(retrieved?.content).toBe('Updated content');
    });

    it('should create domain when storing resource in new domain', async () => {
      const resourceInNewDomain = {
        ...sampleResource,
        domain: 'new-domain'
      };
      
      await storage.storeResource(resourceInNewDomain);
      
      const domains = await storage.listDomains();
      const newDomain = domains.find(d => d.name === 'new-domain');
      
      expect(newDomain).toBeDefined();
      expect(newDomain?.resourceCount).toBe(1);
    });

    it('should list resources', async () => {
      await storage.storeResource(sampleResource);
      await storage.storeResource({
        ...sampleResource,
        id: 'test-resource-2',
        name: 'Second Resource'
      });
      
      const resources = await storage.listResources();
      
      expect(resources).toHaveLength(2);
      expect(resources.map(r => r.name)).toContain('Test Resource');
      expect(resources.map(r => r.name)).toContain('Second Resource');
    });

    it('should list resources by domain', async () => {
      await storage.storeResource(sampleResource);
      await storage.storeResource({
        ...sampleResource,
        id: 'test-resource-2',
        domain: 'other-domain'
      });
      
      const defaultResources = await storage.listResources('default');
      const otherResources = await storage.listResources('other-domain');
      
      expect(defaultResources).toHaveLength(1);
      expect(otherResources).toHaveLength(1);
      expect(defaultResources[0].domain).toBe('default');
      expect(otherResources[0].domain).toBe('other-domain');
    });

    it('should respect limit and offset', async () => {
      for (let i = 0; i < 5; i++) {
        await storage.storeResource({
          ...sampleResource,
          id: `test-resource-${i}`,
          name: `Resource ${i}`
        });
      }
      
      const page1 = await storage.listResources('default', 2, 0);
      const page2 = await storage.listResources('default', 2, 2);
      
      expect(page1).toHaveLength(2);
      expect(page2).toHaveLength(2);
      expect(page1[0].id).not.toBe(page2[0].id);
    });

    it('should search resources by content', async () => {
      await storage.storeResource(sampleResource);
      await storage.storeResource({
        ...sampleResource,
        id: 'searchable-resource',
        content: 'This resource contains unique searchable content'
      });
      
      const results = await storage.searchResources('searchable');
      
      expect(results).toHaveLength(1);
      expect(results[0].id).toBe('searchable-resource');
    });

    it('should search resources by name and description', async () => {
      await storage.storeResource({
        ...sampleResource,
        name: 'Unique Resource Name',
        description: 'Contains special keyword'
      });
      
      const nameResults = await storage.searchResources('Unique');
      const descResults = await storage.searchResources('special');
      
      expect(nameResults).toHaveLength(1);
      expect(descResults).toHaveLength(1);
    });

    it('should filter search by domain', async () => {
      await storage.storeResource({
        ...sampleResource,
        domain: 'domain1',
        content: 'searchable content'
      });
      await storage.storeResource({
        ...sampleResource,
        id: 'resource-2',
        domain: 'domain2',
        content: 'searchable content'
      });
      
      const domain1Results = await storage.searchResources('searchable', 'domain1');
      const domain2Results = await storage.searchResources('searchable', 'domain2');
      
      expect(domain1Results).toHaveLength(1);
      expect(domain2Results).toHaveLength(1);
      expect(domain1Results[0].domain).toBe('domain1');
      expect(domain2Results[0].domain).toBe('domain2');
    });
  });

  describe('Prompt Management', () => {
    const samplePrompt: Omit<PromptTemplate, 'createdAt' | 'updatedAt'> = {
      id: 'test-prompt-1',
      name: 'Test Prompt',
      description: 'A test prompt template',
      template: 'Hello {{name}}, welcome to {{app}}!',
      variables: ['name', 'app'],
      domain: 'default'
    };

    it('should store a prompt template', async () => {
      await storage.storePrompt(samplePrompt);
      
      const retrieved = await storage.getPrompt('test-prompt-1');
      
      expect(retrieved).toBeDefined();
      expect(retrieved?.name).toBe(samplePrompt.name);
      expect(retrieved?.template).toBe(samplePrompt.template);
      expect(retrieved?.variables).toEqual(samplePrompt.variables);
      expect(retrieved?.createdAt).toBeInstanceOf(Date);
    });

    it('should update existing prompt', async () => {
      await storage.storePrompt(samplePrompt);
      
      const updatedPrompt = {
        ...samplePrompt,
        template: 'Updated template with {{newVar}}',
        variables: ['newVar']
      };
      
      await storage.storePrompt(updatedPrompt);
      const retrieved = await storage.getPrompt('test-prompt-1');
      
      expect(retrieved?.template).toBe('Updated template with {{newVar}}');
      expect(retrieved?.variables).toEqual(['newVar']);
    });

    it('should list prompts', async () => {
      await storage.storePrompt(samplePrompt);
      await storage.storePrompt({
        ...samplePrompt,
        id: 'test-prompt-2',
        name: 'Second Prompt'
      });
      
      const prompts = await storage.listPrompts();
      
      expect(prompts).toHaveLength(2);
      expect(prompts.map(p => p.name)).toContain('Test Prompt');
      expect(prompts.map(p => p.name)).toContain('Second Prompt');
    });

    it('should list prompts by domain', async () => {
      await storage.storePrompt(samplePrompt);
      await storage.storePrompt({
        ...samplePrompt,
        id: 'test-prompt-2',
        domain: 'other-domain'
      });
      
      const defaultPrompts = await storage.listPrompts('default');
      const otherPrompts = await storage.listPrompts('other-domain');
      
      expect(defaultPrompts).toHaveLength(1);
      expect(otherPrompts).toHaveLength(1);
      expect(defaultPrompts[0].domain).toBe('default');
      expect(otherPrompts[0].domain).toBe('other-domain');
    });

    it('should create domain when storing prompt in new domain', async () => {
      const promptInNewDomain = {
        ...samplePrompt,
        domain: 'prompt-domain'
      };
      
      await storage.storePrompt(promptInNewDomain);
      
      const domains = await storage.listDomains();
      const newDomain = domains.find(d => d.name === 'prompt-domain');
      
      expect(newDomain).toBeDefined();
    });
  });
});