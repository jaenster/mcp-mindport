# MindPort MCP Server

A high-performance Model Context Protocol (MCP) server built in TypeScript/Node.js, designed for AI systems to efficiently store, search, and retrieve resources with minimal token usage and modern search capabilities.

## ✨ Features

- **🔍 Advanced Search**: Token-efficient fuzzy search, regex patterns, and grep-like functionality
- **📦 SQLite Storage**: Reliable, lightweight database with domain isolation
- **🏷️ Smart Organization**: Domain-based resource management with tag filtering
- **🎯 AI-Optimized**: Designed specifically for Claude Desktop/Code integration
- **⚡ High Performance**: Fast search and retrieval optimized for large datasets  
- **🧪 Comprehensive Testing**: 76+ tests covering all functionality
- **📝 Prompt Templates**: Store and render reusable prompt templates with variables

## 🚀 Quick Start

### For Claude Desktop Users (Recommended)

```bash
# Install dependencies
npm install

# Configure Claude Desktop integration  
npm run start -- --list-domains  # Test installation

# MindPort is now available in Claude Desktop!
```

### Manual Setup

```bash
# Clone repository
git clone <repository>
cd mcp-mindport

# Install dependencies
npm install

# Run server in stdio mode (for MCP clients)
npm start

# Run with custom domain
npm start -- --domain my-project

# List available domains
npm start -- --list-domains
```

### Development & Testing

```bash
# Run comprehensive test suite (76 tests)
npm test

# Run tests once
npm run test:run

# Run tests with UI
npm run test:ui

# Run with coverage
npm run test:coverage
```

## 🔧 Configuration

The server automatically creates configuration at `~/.config/mindport/config.yaml`:

```yaml
server:
  host: "localhost"
  port: 8080

storage:
  path: "~/.config/mindport/data/storage.db"

search:
  index_path: "~/.config/mindport/data/search"

domain:
  default_domain: "default"
```

### Environment Variables

```bash
# Disable logging (recommended for MCP)
export MCP_MINDPORT_LOG=discard

# Set custom domain
export MCP_MINDPORT_DOMAIN=my-project

# Custom storage path
export MCP_MINDPORT_STORE_PATH=/path/to/storage.db
```

## 🛠️ Claude Desktop Integration

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mindport": {
      "command": "npx",
      "args": ["ts-node", "/path/to/mcp-mindport/index.ts"],
      "env": {
        "MCP_MINDPORT_LOG": "discard"
      }
    }
  }
}
```

**Restart Claude Desktop** to activate MindPort!

## 📚 Available MCP Tools

### Resource Management

#### `store_resource`
Store content with metadata and tags
```typescript
{
  id: "api-docs-v1",
  name: "API Documentation", 
  description: "REST API endpoints and authentication",
  content: "GET /users - Retrieve users...",
  tags: ["api", "documentation", "rest"],
  mimeType: "text/markdown"
}
```

#### `get_resource`
Retrieve specific resource by ID
```typescript
{ id: "api-docs-v1" }
```

#### `list_resources` 
List resources in current domain
```typescript
{ limit: 20, offset: 0 }
```

### Search & Discovery

#### `search_resources`
Fast, token-efficient fuzzy search
```typescript
{
  query: "API authentication methods",
  limit: 10
}
```

#### `advanced_search`
Complex queries with tag filtering
```typescript
{
  query: "database design",
  tags: ["sql", "performance"],
  exactTags: true
}
```

#### `grep`
Regex pattern matching (like ripgrep)
```typescript
{ pattern: "function\\s+\\w+\\(" }
```

#### `find`
Find resources by name patterns
```typescript
{ pattern: "^API.*" }
```

### Domain Management

#### `list_domains`
List all available domains
```typescript
{}
```

#### `create_domain` 
Create new domain context
```typescript
{
  name: "frontend-project",
  description: "Frontend development resources"
}
```

#### `switch_domain`
Change current domain
```typescript
{ domain: "frontend-project" }
```

#### `domain_stats`
Get domain statistics and top tags
```typescript
{ domain: "frontend-project" }  // optional
```

### Prompt Templates

#### `store_prompt`
Store reusable prompt templates
```typescript
{
  id: "code-review",
  name: "Code Review Prompt",
  template: "Review this {{language}} code for {{focus}}:\n\n```{{language}}\n{{code}}\n```",
  variables: ["language", "focus", "code"]
}
```

#### `list_prompts`
List available prompt templates
```typescript
{}
```

#### `get_prompt`
Retrieve and render prompts with variables
```typescript
{
  id: "code-review",
  variables: {
    "language": "TypeScript",
    "focus": "performance",
    "code": "const result = await fetch('/api');"
  }
}
```

## 🏗️ Architecture

```
┌─────────────────┐
│   Claude Desktop │ (MCP Client)
│                 │
└─────────┬───────┘
          │ JSON-RPC 2.0 via stdio
          │
┌─────────▼───────┐
│   MCP Server    │ (TypeScript/Node.js)
│                 │
├─────────────────┤
│ Domain Manager  │ (Project isolation)
│                 │  
├─────────────────┤
│ Fuse.js Search  │ (Fuzzy + pattern search)
│                 │
├─────────────────┤
│ SQLite Storage  │ (Resources + prompts)
│                 │
└─────────────────┘
```

### Key Components

- **TypeScript/Node.js**: Modern, maintainable codebase
- **SQLite**: Reliable embedded database with ACID transactions
- **Fuse.js**: Advanced fuzzy search with scoring and highlighting
- **Official MCP SDK**: Anthropic's official Model Context Protocol implementation
- **Commander.js**: Robust CLI interface with comprehensive options
- **Vitest**: Modern testing framework with 76+ comprehensive tests

## 🎯 Token Optimization

MindPort is specifically optimized for AI interactions:

- **Compact Responses**: Minimal formatting, maximum information density
- **Smart Truncation**: Long content is intelligently summarized
- **Relevance Scoring**: Results ranked by relevance to save tokens
- **Configurable Limits**: Control response size with limit parameters
- **Context-Aware**: Domain isolation reduces noise in search results

## 🔍 Search Capabilities

### Fuzzy Search
```bash
# Finds "JavaScript Tutorial" even with typos
search_resources: "javascrpt tutorial"
```

### Regex Patterns  
```bash
# Find all function definitions
grep: "function\\s+\\w+\\("

# Find resources starting with "API"
find: "^API.*"
```

### Tag-Based Filtering
```bash
# Exact tag matching
advanced_search: { query: "auth", tags: ["security"], exactTags: true }

# Partial tag matching  
advanced_search: { query: "auth", tags: ["sec"], exactTags: false }
```

## 📊 Performance

Tested with 100+ resources:
- **Storage**: < 10s for 100 resources
- **Indexing**: < 1s for search index updates
- **Search**: < 100ms for fuzzy search queries
- **Grep**: < 100ms for regex pattern matching
- **Pagination**: < 500ms for large result sets

## 🧪 Testing

Comprehensive test suite with 76 tests covering:

- **Storage Layer** (17 tests): SQLite operations, CRUD, domain isolation
- **Search Engine** (30 tests): Fuzzy search, patterns, grep, tag filtering
- **MCP Server** (25 tests): All tools, error handling, tool schemas
- **Integration** (4 tests): End-to-end workflows, performance, multi-domain

```bash
# Run all tests
npm test

# Run specific test suites
npm test storage
npm test search  
npm test server
npm test integration
```

## 🚀 Advanced Usage

### Multi-Domain Workflow
```bash
# Create project domains
create_domain: { name: "frontend", description: "Frontend code and docs" }
create_domain: { name: "backend", description: "API and database" }

# Switch contexts
switch_domain: { domain: "frontend" }

# Store domain-specific resources
store_resource: { 
  name: "React Component", 
  content: "const Button = ...",
  tags: ["react", "component"]
}
```

### Template-Driven Prompts
```bash
# Store template
store_prompt: {
  id: "bug-report",
  template: "Bug in {{component}}:\n**Expected:** {{expected}}\n**Actual:** {{actual}}"
}

# Use template
get_prompt: {
  id: "bug-report", 
  variables: { 
    component: "Login Form",
    expected: "User logged in",
    actual: "Error 401"
  }
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `npm run test:run`
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

---

**Built for Claude Desktop** 🤖 | **Optimized for AI Workflows** ⚡ | **TypeScript + SQLite + Vitest** 🛠️