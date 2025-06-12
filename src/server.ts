import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  Tool,
} from '@modelcontextprotocol/sdk/types.js';
import { SQLiteStorage } from './storage.js';
import { FuseSearch } from './search.js';
import { Config, Resource } from './types.js';

export class MCPServer {
  private server: Server;
  private storage: SQLiteStorage;
  private search: FuseSearch;
  private config: Config;

  constructor(storage: SQLiteStorage, search: FuseSearch, config: Config) {
    this.storage = storage;
    this.search = search;
    this.config = config;
    
    this.server = new Server(
      {
        name: 'mcp-mindport',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private setupHandlers(): void {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: this.getTools(),
      };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      try {
        switch (name) {
          case 'store_resource':
            return await this.handleStoreResource(args);
          case 'get_resource':
            return await this.handleGetResource(args);
          case 'list_resources':
            return await this.handleListResources(args);
          case 'search_resources':
            return await this.handleSearchResources(args);
          case 'advanced_search':
            return await this.handleAdvancedSearch(args);
          case 'grep':
            return await this.handleGrep(args);
          case 'find':
            return await this.handleFind(args);
          case 'list_domains':
            return await this.handleListDomains(args);
          case 'create_domain':
            return await this.handleCreateDomain(args);
          case 'switch_domain':
            return await this.handleSwitchDomain(args);
          case 'domain_stats':
            return await this.handleDomainStats(args);
          case 'store_prompt':
            return await this.handleStorePrompt(args);
          case 'list_prompts':
            return await this.handleListPrompts(args);
          case 'get_prompt':
            return await this.handleGetPrompt(args);
          default:
            throw new Error(`Unknown tool: ${name}`);
        }
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
        };
      }
    });
  }

  private getTools(): Tool[] {
    return [
      {
        name: 'store_resource',
        description: 'Store content with metadata and tags',
        inputSchema: {
          type: 'object',
          properties: {
            id: { type: 'string', description: 'Unique identifier for the resource' },
            name: { type: 'string', description: 'Name of the resource' },
            description: { type: 'string', description: 'Description of the resource' },
            content: { type: 'string', description: 'Content to store' },
            tags: { type: 'array', items: { type: 'string' }, description: 'Tags for categorization' },
            mimeType: { type: 'string', description: 'MIME type of the content' },
            uri: { type: 'string', description: 'URI of the resource' },
          },
          required: ['id', 'name', 'content'],
        },
      },
      {
        name: 'get_resource',
        description: 'Retrieve a specific resource by ID',
        inputSchema: {
          type: 'object',
          properties: {
            id: { type: 'string', description: 'Resource ID to retrieve' },
          },
          required: ['id'],
        },
      },
      {
        name: 'list_resources',
        description: 'List resources in current domain',
        inputSchema: {
          type: 'object',
          properties: {
            limit: { type: 'number', description: 'Maximum number of resources to return' },
            offset: { type: 'number', description: 'Number of resources to skip' },
          },
        },
      },
      {
        name: 'search_resources',
        description: 'Fast, token-efficient fuzzy search across resources',
        inputSchema: {
          type: 'object',
          properties: {
            query: { type: 'string', description: 'Search query' },
            limit: { type: 'number', description: 'Maximum results to return', default: 10 },
          },
          required: ['query'],
        },
      },
      {
        name: 'advanced_search',
        description: 'Complex queries with tag filtering',
        inputSchema: {
          type: 'object',
          properties: {
            query: { type: 'string', description: 'Search query' },
            tags: { type: 'array', items: { type: 'string' }, description: 'Filter by tags' },
            exactTags: { type: 'boolean', description: 'Exact tag matching', default: false },
          },
          required: ['query'],
        },
      },
      {
        name: 'grep',
        description: 'Regex pattern matching across content',
        inputSchema: {
          type: 'object',
          properties: {
            pattern: { type: 'string', description: 'Regex pattern to search for' },
          },
          required: ['pattern'],
        },
      },
      {
        name: 'find',
        description: 'Find resources by name/title patterns',
        inputSchema: {
          type: 'object',
          properties: {
            pattern: { type: 'string', description: 'Pattern to match in resource names' },
          },
          required: ['pattern'],
        },
      },
      {
        name: 'list_domains',
        description: 'List available domains',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'create_domain',
        description: 'Create new domain context',
        inputSchema: {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Domain name' },
            description: { type: 'string', description: 'Domain description' },
          },
          required: ['name'],
        },
      },
      {
        name: 'switch_domain',
        description: 'Change current domain',
        inputSchema: {
          type: 'object',
          properties: {
            domain: { type: 'string', description: 'Domain to switch to' },
          },
          required: ['domain'],
        },
      },
      {
        name: 'domain_stats',
        description: 'Get domain statistics',
        inputSchema: {
          type: 'object',
          properties: {
            domain: { type: 'string', description: 'Domain name (optional, uses current domain if not specified)' },
          },
        },
      },
      {
        name: 'store_prompt',
        description: 'Store reusable prompt templates',
        inputSchema: {
          type: 'object',
          properties: {
            id: { type: 'string', description: 'Prompt ID' },
            name: { type: 'string', description: 'Prompt name' },
            description: { type: 'string', description: 'Prompt description' },
            template: { type: 'string', description: 'Prompt template with variables' },
            variables: { type: 'array', items: { type: 'string' }, description: 'Template variables' },
          },
          required: ['id', 'name', 'template'],
        },
      },
      {
        name: 'list_prompts',
        description: 'List available prompts',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'get_prompt',
        description: 'Retrieve and render prompts',
        inputSchema: {
          type: 'object',
          properties: {
            id: { type: 'string', description: 'Prompt ID' },
            variables: { type: 'object', description: 'Variables to substitute in template' },
          },
          required: ['id'],
        },
      },
    ];
  }

  private async refreshSearchIndex(): Promise<void> {
    const resources = await this.storage.listResources(this.config.domain.currentDomain);
    this.search.updateIndex(resources);
  }

  private async handleStoreResource(args: any): Promise<any> {
    const resource: Omit<Resource, 'createdAt' | 'updatedAt'> = {
      id: args.id,
      name: args.name,
      description: args.description,
      content: args.content,
      tags: args.tags || [],
      domain: this.config.domain.currentDomain,
      mimeType: args.mimeType,
      uri: args.uri,
    };

    await this.storage.storeResource(resource);
    await this.refreshSearchIndex();

    return {
      content: [
        {
          type: 'text',
          text: `Resource "${resource.name}" stored successfully in domain "${resource.domain}"`,
        },
      ],
    };
  }

  private async handleGetResource(args: any): Promise<any> {
    const resource = await this.storage.getResource(args.id, this.config.domain.currentDomain);
    
    if (!resource) {
      return {
        content: [
          {
            type: 'text',
            text: `Resource with ID "${args.id}" not found in current domain`,
          },
        ],
      };
    }

    return {
      content: [
        {
          type: 'text',
          text: `**${resource.name}**\n\n${resource.description || ''}\n\n${resource.content}`,
        },
      ],
    };
  }

  private async handleListResources(args: any): Promise<any> {
    const resources = await this.storage.listResources(
      this.config.domain.currentDomain,
      args.limit,
      args.offset
    );

    if (resources.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: 'No resources found in current domain',
          },
        ],
      };
    }

    const list = resources
      .map(r => `• **${r.name}** (${r.id})\n  ${r.description || 'No description'}\n  Tags: ${r.tags.join(', ') || 'none'}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Resources in ${this.config.domain.currentDomain}:**\n\n${list}`,
        },
      ],
    };
  }

  private async handleSearchResources(args: any): Promise<any> {
    await this.refreshSearchIndex();
    const results = this.search.search(args.query, args.limit || 10);

    if (results.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: `No results found for "${args.query}"`,
          },
        ],
      };
    }

    const list = results
      .map(r => `• **${r.resource.name}** (score: ${r.score.toFixed(2)})\n  ${r.resource.description || 'No description'}\n  Matches: ${r.matches.join(', ')}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Search results for "${args.query}":**\n\n${list}`,
        },
      ],
    };
  }

  private async handleAdvancedSearch(args: any): Promise<any> {
    await this.refreshSearchIndex();
    
    let results = this.search.search(args.query, 50);
    
    if (args.tags && args.tags.length > 0) {
      const tagFilteredResources = this.search.searchByTags(args.tags, args.exactTags || false);
      results = results.filter(r => 
        tagFilteredResources.some(tr => tr.id === r.resource.id)
      );
    }

    if (results.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: `No results found for advanced search`,
          },
        ],
      };
    }

    const list = results
      .slice(0, 10)
      .map(r => `• **${r.resource.name}** (score: ${r.score.toFixed(2)})\n  Tags: ${r.resource.tags.join(', ')}\n  ${r.resource.description || 'No description'}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Advanced search results:**\n\n${list}`,
        },
      ],
    };
  }

  private async handleGrep(args: any): Promise<any> {
    await this.refreshSearchIndex();
    const results = this.search.grep(args.pattern);

    if (results.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: `No matches found for pattern "${args.pattern}"`,
          },
        ],
      };
    }

    const list = results
      .slice(0, 10)
      .map(r => `• **${r.resource.name}**\n  Matches: ${r.matches.slice(0, 3).join(', ')}${r.matches.length > 3 ? '...' : ''}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Pattern matches for "${args.pattern}":**\n\n${list}`,
        },
      ],
    };
  }

  private async handleFind(args: any): Promise<any> {
    await this.refreshSearchIndex();
    const results = this.search.findByPattern(args.pattern, 'name');

    if (results.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: `No resources found matching name pattern "${args.pattern}"`,
          },
        ],
      };
    }

    const list = results
      .map(r => `• **${r.name}** (${r.id})\n  ${r.description || 'No description'}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Resources matching "${args.pattern}":**\n\n${list}`,
        },
      ],
    };
  }

  private async handleListDomains(args: any): Promise<any> {
    const domains = await this.storage.listDomains();
    
    const list = domains
      .map(d => `• **${d.name}** (${d.resourceCount} resources)\n  ${d.description || 'No description'}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Available domains:**\n\n${list}\n\n*Current domain: ${this.config.domain.currentDomain}*`,
        },
      ],
    };
  }

  private async handleCreateDomain(args: any): Promise<any> {
    await this.storage.createDomain(args.name, args.description);
    
    return {
      content: [
        {
          type: 'text',
          text: `Domain "${args.name}" created successfully`,
        },
      ],
    };
  }

  private async handleSwitchDomain(args: any): Promise<any> {
    const domains = await this.storage.listDomains();
    const domainExists = domains.some(d => d.name === args.domain);
    
    if (!domainExists) {
      return {
        content: [
          {
            type: 'text',
            text: `Domain "${args.domain}" does not exist`,
          },
        ],
      };
    }

    this.config.domain.currentDomain = args.domain;
    
    return {
      content: [
        {
          type: 'text',
          text: `Switched to domain "${args.domain}"`,
        },
      ],
    };
  }

  private async handleDomainStats(args: any): Promise<any> {
    const domain = args.domain || this.config.domain.currentDomain;
    const resources = await this.storage.listResources(domain);
    const prompts = await this.storage.listPrompts(domain);
    
    const tagCounts: Record<string, number> = {};
    resources.forEach(r => {
      r.tags.forEach(tag => {
        tagCounts[tag] = (tagCounts[tag] || 0) + 1;
      });
    });

    const topTags = Object.entries(tagCounts)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 10)
      .map(([tag, count]) => `${tag}: ${count}`)
      .join(', ');

    const stats = [
      `**Domain:** ${domain}`,
      `**Resources:** ${resources.length}`,
      `**Prompts:** ${prompts.length}`,
      `**Top Tags:** ${topTags || 'none'}`,
    ].join('\n');

    return {
      content: [
        {
          type: 'text',
          text: stats,
        },
      ],
    };
  }

  private async handleStorePrompt(args: any): Promise<any> {
    const prompt = {
      id: args.id,
      name: args.name,
      description: args.description,
      template: args.template,
      variables: args.variables || [],
      domain: this.config.domain.currentDomain,
    };

    await this.storage.storePrompt(prompt);

    return {
      content: [
        {
          type: 'text',
          text: `Prompt "${prompt.name}" stored successfully`,
        },
      ],
    };
  }

  private async handleListPrompts(args: any): Promise<any> {
    const prompts = await this.storage.listPrompts(this.config.domain.currentDomain);

    if (prompts.length === 0) {
      return {
        content: [
          {
            type: 'text',
            text: 'No prompts found in current domain',
          },
        ],
      };
    }

    const list = prompts
      .map(p => `• **${p.name}** (${p.id})\n  ${p.description || 'No description'}\n  Variables: ${p.variables.join(', ') || 'none'}`)
      .join('\n\n');

    return {
      content: [
        {
          type: 'text',
          text: `**Prompts in ${this.config.domain.currentDomain}:**\n\n${list}`,
        },
      ],
    };
  }

  private async handleGetPrompt(args: any): Promise<any> {
    const prompt = await this.storage.getPrompt(args.id, this.config.domain.currentDomain);
    
    if (!prompt) {
      return {
        content: [
          {
            type: 'text',
            text: `Prompt with ID "${args.id}" not found`,
          },
        ],
      };
    }

    let renderedTemplate = prompt.template;
    
    if (args.variables) {
      Object.entries(args.variables).forEach(([key, value]) => {
        const regex = new RegExp(`{{\\s*${key}\\s*}}`, 'g');
        renderedTemplate = renderedTemplate.replace(regex, String(value));
      });
    }

    return {
      content: [
        {
          type: 'text',
          text: `**${prompt.name}**\n\n${renderedTemplate}`,
        },
      ],
    };
  }

  async start(): Promise<void> {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
  }

  async close(): Promise<void> {
    await this.storage.close();
    await this.server.close();
  }
}