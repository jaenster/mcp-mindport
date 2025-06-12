import { describe, it, expect, beforeEach } from 'vitest';
import { FuseSearch } from '../src/search.js';
import { Resource } from '../src/types.js';

describe('FuseSearch', () => {
  let search: FuseSearch;
  let sampleResources: Resource[];

  beforeEach(() => {
    search = new FuseSearch();
    
    sampleResources = [
      {
        id: 'resource-1',
        name: 'JavaScript Tutorial',
        description: 'Learn JavaScript fundamentals',
        content: 'JavaScript is a programming language used for web development. It supports functions, objects, and async programming.',
        tags: ['javascript', 'tutorial', 'programming'],
        domain: 'default',
        mimeType: 'text/markdown',
        createdAt: new Date('2023-01-01'),
        updatedAt: new Date('2023-01-01')
      },
      {
        id: 'resource-2',
        name: 'Python Guide',
        description: 'Python programming basics',
        content: 'Python is a versatile programming language known for its simplicity and readability. Great for beginners.',
        tags: ['python', 'guide', 'programming'],
        domain: 'default',
        mimeType: 'text/markdown',
        createdAt: new Date('2023-01-02'),
        updatedAt: new Date('2023-01-02')
      },
      {
        id: 'resource-3',
        name: 'Database Design',
        description: 'Learn database design principles',
        content: 'Database design is crucial for efficient data storage and retrieval. Covers normalization, indexing, and relationships.',
        tags: ['database', 'design', 'sql'],
        domain: 'default',
        mimeType: 'text/markdown',
        createdAt: new Date('2023-01-03'),
        updatedAt: new Date('2023-01-03')
      },
      {
        id: 'resource-4',
        name: 'React Components',
        description: 'Building reusable React components',
        content: 'React components are the building blocks of React applications. Learn to create functional and class components.',
        tags: ['react', 'javascript', 'components'],
        domain: 'frontend',
        mimeType: 'text/markdown',
        createdAt: new Date('2023-01-04'),
        updatedAt: new Date('2023-01-04')
      }
    ];

    search.updateIndex(sampleResources);
  });

  describe('Fuzzy Search', () => {
    it('should find resources by name', () => {
      const results = search.search('JavaScript');
      
      expect(results).toHaveLength(2);
      expect(results[0].resource.name).toBe('JavaScript Tutorial');
      expect(results[1].resource.name).toBe('React Components');
    });

    it('should find resources by description', () => {
      const results = search.search('programming basics');
      
      expect(results.length).toBeGreaterThan(0);
      const pythonResult = results.find(r => r.resource.name === 'Python Guide');
      expect(pythonResult).toBeDefined();
    });

    it('should find resources by content', () => {
      const results = search.search('versatile programming language');
      
      expect(results.length).toBeGreaterThan(0);
      const pythonResult = results.find(r => r.resource.name === 'Python Guide');
      expect(pythonResult).toBeDefined();
    });

    it('should find resources by tags', () => {
      const results = search.search('database');
      
      expect(results.length).toBeGreaterThan(0);
      const dbResult = results.find(r => r.resource.name === 'Database Design');
      expect(dbResult).toBeDefined();
    });

    it('should return results with scores', () => {
      const results = search.search('JavaScript');
      
      results.forEach(result => {
        expect(result.score).toBeGreaterThan(0);
        expect(result.score).toBeLessThanOrEqual(1);
      });
    });

    it('should return results with matches', () => {
      const results = search.search('JavaScript');
      
      results.forEach(result => {
        expect(Array.isArray(result.matches)).toBe(true);
      });
    });

    it('should respect limit parameter', () => {
      const results = search.search('programming', 1);
      
      expect(results).toHaveLength(1);
    });

    it('should handle empty queries gracefully', () => {
      const results = search.search('');
      
      expect(results).toHaveLength(0);
    });

    it('should handle queries with no matches', () => {
      const results = search.search('nonexistentquery12345');
      
      expect(results).toHaveLength(0);
    });
  });

  describe('Tag Search', () => {
    it('should find resources by exact tag match', () => {
      const results = search.searchByTags(['javascript'], true);
      
      expect(results).toHaveLength(2);
      expect(results.every(r => r.tags.includes('javascript'))).toBe(true);
    });

    it('should find resources by partial tag match', () => {
      const results = search.searchByTags(['prog'], false);
      
      expect(results.length).toBeGreaterThan(0);
      expect(results.every(r => 
        r.tags.some(tag => tag.toLowerCase().includes('prog'))
      )).toBe(true);
    });

    it('should find resources matching all tags (exact)', () => {
      const results = search.searchByTags(['javascript', 'tutorial'], true);
      
      expect(results).toHaveLength(1);
      expect(results[0].name).toBe('JavaScript Tutorial');
    });

    it('should find resources matching any tag (partial)', () => {
      const results = search.searchByTags(['java', 'python'], false);
      
      expect(results.length).toBeGreaterThan(0);
      const hasJavaScript = results.some(r => r.tags.includes('javascript'));
      const hasPython = results.some(r => r.tags.includes('python'));
      expect(hasJavaScript || hasPython).toBe(true);
    });

    it('should handle empty tag arrays', () => {
      const exactResults = search.searchByTags([], true);
      const partialResults = search.searchByTags([], false);
      
      expect(exactResults).toHaveLength(sampleResources.length);
      expect(partialResults).toHaveLength(0);
    });
  });

  describe('Pattern Search', () => {
    it('should find resources by regex pattern in content', () => {
      const results = search.findByPattern('programming language', 'content');
      
      expect(results.length).toBeGreaterThan(0);
      results.forEach(resource => {
        expect(resource.content.toLowerCase()).toMatch(/programming language/i);
      });
    });

    it('should find resources by pattern in name', () => {
      const results = search.findByPattern('^[JP].*', 'name');
      
      expect(results.length).toBe(2); // JavaScript Tutorial starts with J, Python Guide starts with P
      expect(results.map(r => r.name)).toContain('JavaScript Tutorial');
      expect(results.map(r => r.name)).toContain('Python Guide');
    });

    it('should find resources by pattern in any field', () => {
      const results = search.findByPattern('React');
      
      expect(results.length).toBe(1);
      expect(results[0].name).toBe('React Components');
    });

    it('should handle invalid regex patterns gracefully', () => {
      expect(() => {
        search.findByPattern('[invalid');
      }).toThrow();
    });

    it('should be case insensitive', () => {
      const results = search.findByPattern('JAVASCRIPT');
      
      expect(results.length).toBeGreaterThan(0);
    });
  });

  describe('Grep Search', () => {
    it('should find pattern matches with scores', () => {
      const results = search.grep('programming');
      
      expect(results.length).toBeGreaterThan(0);
      results.forEach(result => {
        expect(result.score).toBeGreaterThan(0);
        expect(result.matches.length).toBeGreaterThan(0);
      });
    });

    it('should find matches in content', () => {
      const results = search.grep('versatile');
      
      expect(results).toHaveLength(1);
      expect(results[0].resource.name).toBe('Python Guide');
      expect(results[0].matches).toContain('versatile');
    });

    it('should find matches in names', () => {
      const results = search.grep('Tutorial');
      
      expect(results).toHaveLength(1);
      expect(results[0].resource.name).toBe('JavaScript Tutorial');
    });

    it('should find matches in descriptions', () => {
      const results = search.grep('basics');
      
      expect(results).toHaveLength(1);
      expect(results[0].resource.name).toBe('Python Guide');
    });

    it('should find matches in tags', () => {
      const results = search.grep('sql');
      
      expect(results).toHaveLength(1);
      expect(results[0].resource.name).toBe('Database Design');
    });

    it('should return multiple matches per resource', () => {
      const results = search.grep('programming');
      
      const jsResult = results.find(r => r.resource.name === 'JavaScript Tutorial');
      if (jsResult) {
        expect(jsResult.matches.length).toBeGreaterThan(1);
      }
    });

    it('should sort results by score', () => {
      const results = search.grep('programming');
      
      for (let i = 1; i < results.length; i++) {
        expect(results[i - 1].score).toBeGreaterThanOrEqual(results[i].score);
      }
    });

    it('should handle regex patterns', () => {
      const results = search.grep('\\b[Jj]ava[Ss]cript\\b');
      
      expect(results.length).toBeGreaterThan(0);
      expect(results[0].resource.name).toBe('JavaScript Tutorial');
    });

    it('should handle empty patterns', () => {
      const results = search.grep('');
      
      expect(results).toHaveLength(0);
    });
  });

  describe('Index Management', () => {
    it('should update search index with new resources', () => {
      const newResource: Resource = {
        id: 'new-resource',
        name: 'New Technology',
        description: 'Latest tech trends',
        content: 'Artificial intelligence and machine learning',
        tags: ['ai', 'ml', 'tech'],
        domain: 'default',
        mimeType: 'text/markdown',
        createdAt: new Date(),
        updatedAt: new Date()
      };

      const updatedResources = [...sampleResources, newResource];
      search.updateIndex(updatedResources);

      const results = search.search('artificial intelligence');
      expect(results).toHaveLength(1);
      expect(results[0].resource.name).toBe('New Technology');
    });

    it('should handle empty resource arrays', () => {
      search.updateIndex([]);
      
      const results = search.search('JavaScript');
      expect(results).toHaveLength(0);
    });
  });
});