# MindPort MCP Server

A high-performance Model Context Protocol (MCP) server built in TypeScript/Node.js that provides optimized storage and search capabilities for AI systems. Designed for seamless integration with Claude Desktop and other MCP clients.

## Features

- **Advanced Search**: Token-efficient fuzzy search, regex patterns, and grep-like functionality
- **SQLite Storage**: Reliable, lightweight database with domain isolation
- **Smart Organization**: Domain-based resource management with tag filtering
- **AI-Optimized**: Designed specifically for Claude Desktop/Code integration
- **High Performance**: Fast search and retrieval optimized for large datasets
- **Comprehensive Testing**: 76+ tests covering all functionality
- **Prompt Templates**: Store and render reusable prompt templates with variables
- **Modern Web Interface**: Professional dashboard for browsing and managing resources

## Installation

### NPM Package

```bash
npm install -g mindport-mcp
mindport --help
```

### From Source

```bash
git clone https://github.com/mindport-ai/mcp-mindport.git
cd mcp-mindport
npm install
npm run build
```

### Production Deployment

```bash
# Install globally
npm install -g mindport-mcp

# Or run with npx
npx mindport-mcp

# Start web interface
npx mindport-mcp --web
```

The installation automatically:
- Installs all dependencies
- Sets up SQLite database
- Configures default settings
- Creates necessary directories

## Quick Start

### Development Mode

```bash
# Install dependencies
npm install

# Start MCP server (for Claude Desktop)
npm run dev

# Start web interface (in new terminal)
npm run web
# Visit http://localhost:3001
```

### Production Mode

```bash
# Build for production
npm run build

# Start production server
npm start

# Start web interface
npm run web:start
```

### Testing

```bash
# Run comprehensive test suite
npm test

# Run tests once
npm run test:run

# Run with coverage
npm run test:coverage
```

## Production Deployment

### Publishing to NPM

```bash
# Prepare for release
npm run build
npm run test:run

# Publish to NPM
npm publish

# Install globally from NPM
npm install -g mindport-mcp
```

### Docker Deployment

```bash
# Build Docker image
docker build -t mindport-mcp .

# Run in container
docker run -d -p 3001:3001 \
  -v ~/.config/mindport:/root/.config/mindport \
  mindport-mcp
```

### Production Server Setup

```bash
# Install globally
npm install -g mindport-mcp

# Create systemd service (Linux)
sudo tee /etc/systemd/system/mindport.service << EOF
[Unit]
Description=MindPort MCP Server
After=network.target

[Service]
Type=simple
User=mindport
WorkingDirectory=/opt/mindport
ExecStart=/usr/bin/node /usr/local/bin/mindport-mcp
Restart=always
Environment=NODE_ENV=production
Environment=MCP_MINDPORT_LOG=/var/log/mindport.log

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl enable mindport
sudo systemctl start mindport
```

## Configuration

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

## Claude Desktop Integration

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mindport": {
      "command": "npx",
      "args": ["mindport-mcp"],
      "env": {
        "MCP_MINDPORT_LOG": "discard"
      }
    }
  }
}
```

**Restart Claude Desktop** to activate MindPort.

## Available MCP Tools

### Resource Management

#### `store_resource`
Store content with metadata and tags
```json5
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
```json5
{ id: "api-docs-v1" }
```

#### `list_resources` 
List resources in current domain
```json5
{ limit: 20, offset: 0 }
```

### Search & Discovery

#### `search_resources`
Fast, token-efficient fuzzy search
```json5
{
  query: "API authentication methods",
  limit: 10
}
```

#### `advanced_search`
Complex queries with tag filtering
```json5
{
  query: "database design",
  tags: ["sql", "performance"],
  exactTags: true
}
```

#### `grep`
Regex pattern matching (like ripgrep)
```json5
{ pattern: "function\\s+\\w+\\(" }
```

#### `find`
Find resources by name patterns
```json5
{ pattern: "^API.*" }
```

### Domain Management

#### `list_domains`
List all available domains
```json5
{}
```

#### `create_domain` 
Create new domain context
```json5
{
  name: "frontend-project",
  description: "Frontend development resources"
}
```

#### `switch_domain`
Change current domain
```json5
{ domain: "frontend-project" }
```

#### `domain_stats`
Get domain statistics and top tags
```json5
{ domain: "frontend-project" }  // optional
```

### Prompt Templates

#### `store_prompt`
Store reusable prompt templates
```json5
{
  id: "code-review",
  name: "Code Review Prompt",
  template: "Review this {{language}} code for {{focus}}:\n\n```{{language}}\n{{code}}\n```",
  variables: ["language", "focus", "code"]
}
```

#### `list_prompts`
List available prompt templates
```json5
{}
```

#### `get_prompt`
Retrieve and render prompts with variables
```json5
{
  id: "code-review",
  variables: {
    "language": "TypeScript",
    "focus": "performance",
    "code": "const result = await fetch('/api');"
  }
}
```

## Architecture

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

## Token Optimization

MindPort is specifically optimized for AI interactions:

- **Compact Responses**: Minimal formatting, maximum information density
- **Smart Truncation**: Long content is intelligently summarized
- **Relevance Scoring**: Results ranked by relevance to save tokens
- **Configurable Limits**: Control response size with limit parameters
- **Context-Aware**: Domain isolation reduces noise in search results

## Search Capabilities

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

## Performance

Tested with 100+ resources:
- **Storage**: < 10s for 100 resources
- **Indexing**: < 1s for search index updates
- **Search**: < 100ms for fuzzy search queries
- **Grep**: < 100ms for regex pattern matching
- **Pagination**: < 500ms for large result sets

## Testing

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

## Advanced Usage

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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `npm run test:run`
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

---

Built for Claude Desktop | Optimized for AI Workflows | TypeScript + SQLite + Vitest