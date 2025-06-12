import Fuse from 'fuse.js';
import { Resource, SearchResult } from './types.js';

export class FuseSearch {
  private fuse: Fuse<Resource>;
  private resources: Resource[] = [];

  constructor() {
    this.fuse = new Fuse(this.resources, {
      keys: [
        { name: 'name', weight: 0.4 },
        { name: 'description', weight: 0.3 },
        { name: 'content', weight: 0.2 },
        { name: 'tags', weight: 0.1 }
      ],
      threshold: 0.4,
      includeScore: true,
      includeMatches: true,
      minMatchCharLength: 2
    });
  }

  updateIndex(resources: Resource[]): void {
    this.resources = resources;
    this.fuse.setCollection(resources);
  }

  search(query: string, limit = 10): SearchResult[] {
    const results = this.fuse.search(query, { limit });
    
    return results.map(result => ({
      resource: result.item,
      score: 1 - (result.score || 0),
      matches: result.matches?.map(match => match.value || '') || []
    }));
  }

  searchByTags(tags: string[], exact = false): Resource[] {
    if (exact) {
      return this.resources.filter(resource => 
        tags.every(tag => resource.tags.includes(tag))
      );
    } else {
      return this.resources.filter(resource => 
        tags.some(tag => resource.tags.some(resourceTag => 
          resourceTag.toLowerCase().includes(tag.toLowerCase())
        ))
      );
    }
  }

  findByPattern(pattern: string, field?: keyof Resource): Resource[] {
    return this.resources.filter(resource => {
      // Create a new regex for each resource to avoid global regex state issues
      const regex = new RegExp(pattern, 'gi');
      
      if (field) {
        const value = resource[field];
        if (typeof value === 'string') {
          return regex.test(value);
        } else if (Array.isArray(value)) {
          return value.some(item => regex.test(String(item)));
        }
        return false;
      } else {
        // Create fresh regex instances for each test to avoid state issues
        const nameRegex = new RegExp(pattern, 'gi');
        const descRegex = new RegExp(pattern, 'gi');
        const contentRegex = new RegExp(pattern, 'gi');
        const tagRegex = new RegExp(pattern, 'gi');
        
        return nameRegex.test(resource.name) || 
               descRegex.test(resource.description || '') ||
               contentRegex.test(resource.content) ||
               resource.tags.some(tag => tagRegex.test(tag));
      }
    });
  }

  grep(pattern: string): SearchResult[] {
    if (!pattern || pattern.trim() === '') {
      return [];
    }
    
    const regex = new RegExp(pattern, 'gi');
    const results: SearchResult[] = [];

    for (const resource of this.resources) {
      const matches: string[] = [];
      
      // Check content for matches
      const contentMatches = resource.content.match(regex);
      if (contentMatches) {
        matches.push(...contentMatches);
      }

      // Check name and description
      if (regex.test(resource.name)) {
        matches.push(resource.name);
      }
      
      if (resource.description && regex.test(resource.description)) {
        matches.push(resource.description);
      }

      // Check tags
      resource.tags.forEach(tag => {
        if (regex.test(tag)) {
          matches.push(tag);
        }
      });

      if (matches.length > 0) {
        results.push({
          resource,
          score: matches.length / 10, // Simple scoring based on match count
          matches
        });
      }
    }

    // Sort by score descending
    return results.sort((a, b) => b.score - a.score);
  }
}