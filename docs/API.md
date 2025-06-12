# MindPort API Documentation

Complete reference for all MCP tools and their schemas.

## Tool Schemas

### Resource Management

#### store_resource

Store content with metadata and tags in the current domain.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    id: { 
      type: "string", 
      description: "Unique identifier for the resource",
      required: true
    },
    name: { 
      type: "string", 
      description: "Human-readable name of the resource",
      required: true
    },
    description: { 
      type: "string", 
      description: "Optional description of the resource content"
    },
    content: { 
      type: "string", 
      description: "The actual content to store",
      required: true
    },
    tags: { 
      type: "array", 
      items: { type: "string" },
      description: "Tags for categorization and filtering"
    },
    mimeType: { 
      type: "string", 
      description: "MIME type of the content (e.g., 'text/markdown')"
    },
    uri: { 
      type: "string", 
      description: "Optional URI reference for the resource"
    }
  },
  required: ["id", "name", "content"]
}
```

**Example:**
```json
{
  "id": "api-auth-guide",
  "name": "API Authentication Guide",
  "description": "Complete guide for API authentication methods",
  "content": "# API Authentication\\n\\nThis guide covers OAuth 2.0, JWT tokens...",
  "tags": ["api", "authentication", "security", "guide"],
  "mimeType": "text/markdown",
  "uri": "https://docs.example.com/auth"
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "Resource \"API Authentication Guide\" stored successfully in domain \"default\""
    }
  ]
}
```

#### get_resource

Retrieve a specific resource by its ID from the current domain.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    id: { 
      type: "string", 
      description: "Resource ID to retrieve",
      required: true
    }
  },
  required: ["id"]
}
```

**Example:**
```json
{
  "id": "api-auth-guide"
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**API Authentication Guide**\\n\\nComplete guide for API authentication methods\\n\\n# API Authentication\\n\\nThis guide covers OAuth 2.0, JWT tokens..."
    }
  ]
}
```

#### list_resources

List resources in the current domain with optional pagination.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    limit: { 
      type: "number", 
      description: "Maximum number of resources to return"
    },
    offset: { 
      type: "number", 
      description: "Number of resources to skip for pagination"
    }
  }
}
```

**Example:**
```json
{
  "limit": 10,
  "offset": 0
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Resources in default:**\\n\\n• **API Authentication Guide** (api-auth-guide)\\n  Complete guide for API authentication methods\\n  Tags: api, authentication, security, guide\\n\\n• **Database Schema** (db-schema-v2)\\n  User and product table definitions\\n  Tags: database, schema, sql"
    }
  ]
}
```

### Search & Discovery

#### search_resources

Perform fuzzy search across all resource content with intelligent ranking.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    query: { 
      type: "string", 
      description: "Search query string",
      required: true
    },
    limit: { 
      type: "number", 
      description: "Maximum results to return",
      default: 10
    }
  },
  required: ["query"]
}
```

**Example:**
```json
{
  "query": "API authentication JWT OAuth",
  "limit": 5
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text", 
      "text": "**Search results for \"API authentication JWT OAuth\":**\\n\\n• **API Authentication Guide** (score: 0.95)\\n  Complete guide for API authentication methods\\n  Matches: API, authentication, JWT, OAuth\\n\\n• **Security Best Practices** (score: 0.78)\\n  Security guidelines and recommendations\\n  Matches: authentication, API"
    }
  ]
}
```

#### advanced_search

Complex search with tag filtering and exact matching options.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    query: { 
      type: "string", 
      description: "Search query string",
      required: true
    },
    tags: { 
      type: "array", 
      items: { type: "string" },
      description: "Filter results by these tags"
    },
    exactTags: { 
      type: "boolean", 
      description: "Require exact tag matching",
      default: false
    }
  },
  required: ["query"]
}
```

**Example:**
```json
{
  "query": "database performance",
  "tags": ["sql", "optimization"],
  "exactTags": true
}
```

#### grep

Search for regex patterns across all resource content (similar to ripgrep).

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    pattern: { 
      type: "string", 
      description: "Regular expression pattern to search for",
      required: true
    }
  },
  required: ["pattern"]
}
```

**Example:**
```json
{
  "pattern": "function\\s+\\w+\\s*\\("
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Pattern matches for \"function\\\\s+\\\\w+\\\\s*\\\\(\":**\\n\\n• **React Components Guide**\\n  Matches: function Button(, function useAuth(, function handleClick(\\n\\n• **API Utils**\\n  Matches: function fetchData(, function parseResponse("
    }
  ]
}
```

#### find

Find resources by name/title pattern matching.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    pattern: { 
      type: "string", 
      description: "Pattern to match in resource names",
      required: true
    }
  },
  required: ["pattern"]
}
```

**Example:**
```json
{
  "pattern": "^API.*"
}
```

### Domain Management

#### list_domains

List all available domains with resource counts.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {}
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Available domains:**\\n\\n• **default** (15 resources)\\n  Default domain for resources\\n\\n• **frontend-project** (8 resources)\\n  Frontend development resources\\n\\n• **backend-api** (12 resources)\\n  Backend API documentation and code\\n\\n*Current domain: default*"
    }
  ]
}
```

#### create_domain

Create a new domain for organizing resources.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    name: { 
      type: "string", 
      description: "Domain name (must be unique)",
      required: true
    },
    description: { 
      type: "string", 
      description: "Optional description of the domain purpose"
    }
  },
  required: ["name"]
}
```

**Example:**
```json
{
  "name": "mobile-app",
  "description": "Resources for mobile application development"
}
```

#### switch_domain

Change the current active domain context.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    domain: { 
      type: "string", 
      description: "Domain name to switch to",
      required: true
    }
  },
  required: ["domain"]
}
```

**Example:**
```json
{
  "domain": "frontend-project"
}
```

#### domain_stats

Get statistics and analytics for a domain.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    domain: { 
      type: "string", 
      description: "Domain name (uses current domain if not specified)"
    }
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Domain:** frontend-project\\n**Resources:** 8\\n**Prompts:** 3\\n**Top Tags:** react: 5, javascript: 4, component: 3, css: 2, testing: 2"
    }
  ]
}
```

### Prompt Templates

#### store_prompt

Store a reusable prompt template with variables.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    id: { 
      type: "string", 
      description: "Unique prompt ID",
      required: true
    },
    name: { 
      type: "string", 
      description: "Human-readable prompt name",
      required: true
    },
    description: { 
      type: "string", 
      description: "Optional description of the prompt purpose"
    },
    template: { 
      type: "string", 
      description: "Prompt template with {{variable}} placeholders",
      required: true
    },
    variables: { 
      type: "array", 
      items: { type: "string" },
      description: "List of variable names used in the template"
    }
  },
  required: ["id", "name", "template"]
}
```

**Example:**
```json
{
  "id": "code-review",
  "name": "Code Review Request",
  "description": "Template for requesting code reviews",
  "template": "Please review this {{language}} code for {{focus}}:\\n\\n```{{language}}\\n{{code}}\\n```\\n\\nPay special attention to {{aspects}}.",
  "variables": ["language", "focus", "code", "aspects"]
}
```

#### list_prompts

List all available prompt templates in the current domain.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {}
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Prompts in default:**\\n\\n• **Code Review Request** (code-review)\\n  Template for requesting code reviews\\n  Variables: language, focus, code, aspects\\n\\n• **Bug Report** (bug-report)\\n  Structured bug reporting template\\n  Variables: component, expected, actual, steps"
    }
  ]
}
```

#### get_prompt

Retrieve and optionally render a prompt template with variable substitution.

**Input Schema:**
```typescript
{
  type: "object",
  properties: {
    id: { 
      type: "string", 
      description: "Prompt ID to retrieve",
      required: true
    },
    variables: { 
      type: "object", 
      description: "Variables to substitute in the template"
    }
  },
  required: ["id"]
}
```

**Example:**
```json
{
  "id": "code-review",
  "variables": {
    "language": "TypeScript",
    "focus": "performance optimization",
    "code": "const processData = (items: any[]) => {\\n  return items.map(item => expensiveOperation(item));\\n};",
    "aspects": "algorithmic complexity and memory usage"
  }
}
```

**Response:**
```json
{
  "content": [
    {
      "type": "text",
      "text": "**Code Review Request**\\n\\nPlease review this TypeScript code for performance optimization:\\n\\n```typescript\\nconst processData = (items: any[]) => {\\n  return items.map(item => expensiveOperation(item));\\n};\\n```\\n\\nPay special attention to algorithmic complexity and memory usage."
    }
  ]
}
```

## Error Handling

All tools return error responses in this format when something goes wrong:

```json
{
  "content": [
    {
      "type": "text",
      "text": "Error: [descriptive error message]"
    }
  ]
}
```

Common error scenarios:
- **Resource not found**: "Resource with ID 'xyz' not found in current domain"
- **Domain not found**: "Domain 'xyz' does not exist"
- **Invalid regex**: "Invalid regular expression pattern"
- **Missing required fields**: "Missing required field: [field name]"

## Response Format

All successful responses follow the MCP standard format:

```typescript
{
  content: Array<{
    type: "text";
    text: string;
  }>;
}
```

The `text` field contains formatted markdown content optimized for readability and token efficiency.

## Performance Characteristics

- **Search Operations**: < 100ms for datasets up to 1000 resources
- **Storage Operations**: < 50ms for individual resource operations
- **Bulk Operations**: < 10s for 100+ resources
- **Memory Usage**: ~50MB baseline + ~1KB per stored resource
- **Concurrent Operations**: Fully thread-safe with SQLite WAL mode

## Rate Limits

No built-in rate limiting. Performance is primarily limited by:
- SQLite I/O operations
- Search index size
- Available system memory
- Node.js event loop capacity