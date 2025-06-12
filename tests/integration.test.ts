import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { promises as fs } from 'fs';
import path from 'path';
import os from 'os';
import { SQLiteStorage } from '../src/storage.js';
import { FuseSearch } from '../src/search.js';
import { Resource } from '../src/types.js';

describe('Integration Tests', () => {
  let storage: SQLiteStorage;
  let search: FuseSearch;
  let testDbPath: string;

  beforeEach(async () => {
    testDbPath = path.join(os.tmpdir(), `mindport-integration-test-${Date.now()}-${Math.random()}.db`);
    storage = new SQLiteStorage(testDbPath);
    await storage.initialize();
    search = new FuseSearch();
  });

  afterEach(async () => {
    await storage.close();
    try {
      await fs.unlink(testDbPath);
    } catch (error) {
      // Ignore cleanup errors
    }
  });

  describe('End-to-End Workflow', () => {
    it('should handle complete resource lifecycle', async () => {
      // 1. Create a domain
      await storage.createDomain('project-alpha', 'Alpha project domain');
      
      // 2. Store resources in the domain
      const resources: Omit<Resource, 'createdAt' | 'updatedAt'>[] = [
        {
          id: 'doc-1',
          name: 'API Documentation',
          description: 'REST API endpoints documentation',
          content: 'GET /users - Retrieve all users\nPOST /users - Create new user\nPUT /users/:id - Update user',
          tags: ['api', 'documentation', 'rest'],
          domain: 'project-alpha',
          mimeType: 'text/markdown',
        },
        {
          id: 'code-1',
          name: 'User Controller',
          description: 'Express.js user controller implementation',
          content: 'const express = require("express");\nconst router = express.Router();\n\nrouter.get("/users", async (req, res) => {\n  // Implementation\n});',
          tags: ['javascript', 'express', 'controller'],
          domain: 'project-alpha',
          mimeType: 'application/javascript',
        },
        {
          id: 'test-1',
          name: 'User Tests',
          description: 'Unit tests for user functionality',
          content: 'describe("User Controller", () => {\n  it("should return all users", async () => {\n    // Test implementation\n  });\n});',
          tags: ['javascript', 'testing', 'jest'],
          domain: 'project-alpha',
          mimeType: 'application/javascript',
        }
      ];

      for (const resource of resources) {
        await storage.storeResource(resource);
      }

      // 3. Update search index
      const allResources = await storage.listResources('project-alpha');
      search.updateIndex(allResources);

      // 4. Verify domain was created and populated
      const domains = await storage.listDomains();
      const alphaDomain = domains.find(d => d.name === 'project-alpha');
      expect(alphaDomain).toBeDefined();
      expect(alphaDomain?.resourceCount).toBe(3);

      // 5. Test various search scenarios
      
      // Fuzzy search for API
      const apiResults = search.search('API');
      expect(apiResults.length).toBeGreaterThan(0);
      expect(apiResults[0].resource.name).toBe('API Documentation');

      // Tag-based search
      const jsResources = search.searchByTags(['javascript'], true);
      expect(jsResources).toHaveLength(2);
      expect(jsResources.map(r => r.name)).toContain('User Controller');
      expect(jsResources.map(r => r.name)).toContain('User Tests');

      // Pattern search in content
      const expressResults = search.findByPattern('express', 'content');
      expect(expressResults).toHaveLength(1);
      expect(expressResults[0].name).toBe('User Controller');

      // Grep search for specific patterns
      const routerMatches = search.grep('router\\.');
      expect(routerMatches.length).toBeGreaterThan(0);
      expect(routerMatches[0].resource.name).toBe('User Controller');

      // 6. Test resource retrieval
      const retrievedDoc = await storage.getResource('doc-1', 'project-alpha');
      expect(retrievedDoc).toBeDefined();
      expect(retrievedDoc?.content).toContain('GET /users');

      // 7. Test cross-domain isolation
      const defaultResources = await storage.listResources('default');
      expect(defaultResources).toHaveLength(0); // Should be empty

      // 8. Update a resource and verify search index stays consistent
      const updatedResource = {
        ...resources[0],
        content: resources[0].content + '\nDELETE /users/:id - Delete user',
        tags: [...resources[0].tags, 'crud'],
      };
      
      await storage.storeResource(updatedResource);
      
      // Refresh search index
      const updatedResources = await storage.listResources('project-alpha');
      search.updateIndex(updatedResources);
      
      // Verify update
      const deleteResults = search.grep('DELETE');
      expect(deleteResults).toHaveLength(1);
      expect(deleteResults[0].resource.tags).toContain('crud');
    });

    it('should handle prompt templates with complex workflows', async () => {
      // 1. Create prompts for different use cases
      const prompts = [
        {
          id: 'code-review',
          name: 'Code Review Prompt',
          description: 'Template for code review requests',
          template: 'Please review this {{language}} code for {{focus}}:\n\n```{{language}}\n{{code}}\n```\n\nPay special attention to {{aspects}}.',
          variables: ['language', 'focus', 'code', 'aspects'],
          domain: 'default',
        },
        {
          id: 'bug-report',
          name: 'Bug Report Template',
          description: 'Template for bug reports',
          template: 'Bug in {{component}}:\n\n**Expected:** {{expected}}\n**Actual:** {{actual}}\n**Steps:** {{steps}}\n\n{{additional}}',
          variables: ['component', 'expected', 'actual', 'steps', 'additional'],
          domain: 'default',
        }
      ];

      for (const prompt of prompts) {
        await storage.storePrompt(prompt);
      }

      // 2. Verify prompts were stored
      const storedPrompts = await storage.listPrompts('default');
      expect(storedPrompts).toHaveLength(2);

      // 3. Test prompt rendering with variables
      const codeReviewPrompt = await storage.getPrompt('code-review');
      expect(codeReviewPrompt).toBeDefined();

      // Simulate rendering (this would be done by the MCP server)
      let rendered = codeReviewPrompt!.template;
      const variables = {
        language: 'JavaScript',
        focus: 'performance optimization',
        code: 'function slowFunction() { /* code */ }',
        aspects: 'algorithmic complexity and memory usage'
      };

      Object.entries(variables).forEach(([key, value]) => {
        const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
        rendered = rendered.replace(regex, value);
      });

      expect(rendered).toContain('JavaScript code for performance optimization');
      expect(rendered).not.toContain('{{');

      // 4. Test prompt updates
      const updatedPrompt = {
        ...prompts[0],
        template: prompts[0].template + '\n\nAdditional context: {{context}}',
        variables: [...prompts[0].variables, 'context'],
      };

      await storage.storePrompt(updatedPrompt);
      
      const retrievedUpdated = await storage.getPrompt('code-review');
      expect(retrievedUpdated?.variables).toContain('context');
      expect(retrievedUpdated?.template).toContain('Additional context');
    });

    it('should handle multi-domain scenarios', async () => {
      // 1. Create multiple domains
      const domains = [
        { name: 'frontend', description: 'Frontend resources' },
        { name: 'backend', description: 'Backend resources' },
        { name: 'devops', description: 'DevOps and infrastructure' }
      ];

      for (const domain of domains) {
        await storage.createDomain(domain.name, domain.description);
      }

      // 2. Store resources in different domains
      const frontendResource = {
        id: 'react-component',
        name: 'Button Component',
        content: 'import React from "react";\n\nconst Button = ({ children, onClick }) => {\n  return <button onClick={onClick}>{children}</button>;\n};',
        tags: ['react', 'component', 'ui'],
        domain: 'frontend',
      };

      const backendResource = {
        id: 'api-handler',
        name: 'API Handler',
        content: 'const handleRequest = async (req, res) => {\n  // Handle API request\n};',
        tags: ['api', 'handler', 'backend'],
        domain: 'backend',
      };

      const devopsResource = {
        id: 'dockerfile',
        name: 'Docker Configuration',
        content: 'FROM node:16\nWORKDIR /app\nCOPY package*.json ./\nRUN npm install',
        tags: ['docker', 'deployment', 'infrastructure'],
        domain: 'devops',
      };

      await storage.storeResource(frontendResource);
      await storage.storeResource(backendResource);
      await storage.storeResource(devopsResource);

      // 3. Test domain isolation
      const frontendResources = await storage.listResources('frontend');
      const backendResources = await storage.listResources('backend');
      const devopsResources = await storage.listResources('devops');

      expect(frontendResources).toHaveLength(1);
      expect(backendResources).toHaveLength(1);
      expect(devopsResources).toHaveLength(1);

      expect(frontendResources[0].name).toBe('Button Component');
      expect(backendResources[0].name).toBe('API Handler');
      expect(devopsResources[0].name).toBe('Docker Configuration');

      // 4. Test cross-domain search when all resources are indexed together
      const allResources = [
        ...frontendResources,
        ...backendResources,
        ...devopsResources
      ];
      search.updateIndex(allResources);

      const nodeResults = search.search('node');
      expect(nodeResults.length).toBeGreaterThan(0);
      // Should find both the Dockerfile (node:16) and potentially others

      const reactResults = search.search('React');
      expect(reactResults).toHaveLength(1);
      expect(reactResults[0].resource.domain).toBe('frontend');

      // 5. Test tag-based filtering across domains
      const allTagResults = search.searchByTags(['api'], false);
      expect(allTagResults).toHaveLength(1);
      expect(allTagResults[0].domain).toBe('backend');

      // 6. Verify domain statistics
      const allDomains = await storage.listDomains();
      expect(allDomains).toHaveLength(4); // default + 3 new domains
      
      const frontendDomain = allDomains.find(d => d.name === 'frontend');
      expect(frontendDomain?.resourceCount).toBe(1);
    });

    it('should handle search performance with large datasets', async () => {
      // 1. Generate a large number of resources
      const resourceCount = 100;
      const resources: Omit<Resource, 'createdAt' | 'updatedAt'>[] = [];

      for (let i = 0; i < resourceCount; i++) {
        resources.push({
          id: `perf-resource-${i}`,
          name: `Performance Test Resource ${i}`,
          description: `This is resource number ${i} for performance testing`,
          content: `This resource contains test content with unique identifier ${i}. It includes various keywords like performance, testing, search, and optimization-${i % 10}.`,
          tags: [
            'performance',
            'testing',
            i % 2 === 0 ? 'even' : 'odd',
            `batch-${Math.floor(i / 10)}`,
            `category-${i % 5}`
          ],
          domain: 'performance-test',
          mimeType: 'text/plain',
        });
      }

      // 2. Store all resources
      const startStore = Date.now();
      for (const resource of resources) {
        await storage.storeResource(resource);
      }
      const storeTime = Date.now() - startStore;

      // 3. Update search index
      const startIndex = Date.now();
      const allResources = await storage.listResources('performance-test');
      search.updateIndex(allResources);
      const indexTime = Date.now() - startIndex;

      expect(allResources).toHaveLength(resourceCount);

      // 4. Test search performance
      const startSearch = Date.now();
      const searchResults = search.search('performance testing', 10);
      const searchTime = Date.now() - startSearch;

      expect(searchResults.length).toBeGreaterThan(0);
      expect(searchResults.length).toBeLessThanOrEqual(10);

      // 5. Test grep performance
      const startGrep = Date.now();
      const grepResults = search.grep('optimization-[0-9]');
      const grepTime = Date.now() - startGrep;

      expect(grepResults.length).toBe(100); // Each resource contains "optimization-X" pattern

      // 6. Test tag search performance
      const startTagSearch = Date.now();
      const tagResults = search.searchByTags(['even'], true);
      const tagSearchTime = Date.now() - startTagSearch;

      expect(tagResults).toHaveLength(50); // Half of the resources

      // Performance expectations (these are generous bounds)
      expect(storeTime).toBeLessThan(10000); // 10 seconds for 100 resources
      expect(indexTime).toBeLessThan(1000);  // 1 second to index
      expect(searchTime).toBeLessThan(100);  // 100ms for search
      expect(grepTime).toBeLessThan(100);    // 100ms for grep
      expect(tagSearchTime).toBeLessThan(100); // 100ms for tag search

      // 7. Test pagination performance
      const startPagination = Date.now();
      const page1 = await storage.listResources('performance-test', 20, 0);
      const page2 = await storage.listResources('performance-test', 20, 20);
      const paginationTime = Date.now() - startPagination;

      expect(page1).toHaveLength(20);
      expect(page2).toHaveLength(20);
      expect(page1[0].id).not.toBe(page2[0].id);
      expect(paginationTime).toBeLessThan(500); // 500ms for pagination
    });
  });
});