# MindPort

A high-performance Model Context Protocol (MCP) resource server built in Go, designed for AI systems to efficiently store, search, and retrieve data with minimal token usage.

## Features

- **Optimized Search**: Token-efficient search mechanism specifically designed for AI systems
- **Dual Storage**: Fast key-value storage (BadgerDB) combined with full-text search (Bleve)
- **Resource Management**: Store and retrieve various types of resources with metadata
- **Prompt Templates**: Store and manage reusable prompt templates
- **Daemon Mode**: Run as background service with WebSocket support
- **MCP Compatible**: Full Model Context Protocol implementation

## Quick Start

### For Claude Code CLI Users

```bash
# Quick setup (recommended)
./claude-code-setup.sh

# Start daemon
mcp-mindport --daemon &

# Claude Code will automatically use MindPort via environment variables
```

### For Claude Desktop Users

```bash
# Full installation with Claude Desktop integration
./install.sh

# Follow prompts to configure Claude Desktop
# Restart Claude Desktop to load MindPort
```

### Manual Build and Run

```bash
# Clone and build
git clone <repository>
cd mcp-mindport
go mod tidy
go build -o mcp-mindport

# Run in stdio mode (for MCP clients)
./mcp-mindport

# Run as daemon
./mcp-mindport --daemon

# Use custom config
./mcp-mindport --config /path/to/config.yaml
```

### Configuration

Create a `.mcp-mindport.yaml` file in your home directory or current directory:

```yaml
server:
  host: "localhost"
  port: 8080

storage:
  path: "./data/storage"

search:
  index_path: "./data/search"

daemon:
  pid_file: "/tmp/mcp-mindport.pid"
  log_file: "/tmp/mcp-mindport.log"
```

## MCP Integration

### With Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "mindport": {
      "command": "/path/to/mcp-mindport",
      "args": [],
      "env": {}
    }
  }
}
```

### With Other MCP Clients

The server supports both stdio and WebSocket transports:

- **Stdio**: Default mode, use the binary directly
- **WebSocket**: Run with `--daemon` flag, connect to `ws://localhost:8080/mcp`

## Available Tools

### store_resource

Store a new resource with content and metadata.

```json
{
  "title": "API Documentation",
  "content": "REST API endpoints...",
  "type": "documentation",
  "tags": ["api", "docs"],
  "metadata": {"version": "1.0"}
}
```

### search_resources

Search for resources using optimized token-efficient search.

```json
{
  "query": "API authentication",
  "limit": 10,
  "type": "resource",
  "tags": ["api"]
}
```

### store_prompt

Store a reusable prompt template.

```json
{
  "name": "code_review",
  "description": "Code review prompt template",
  "template": "Review this {{language}} code for {{focus}}:\n\n{{code}}",
  "variables": {
    "language": "Programming language",
    "focus": "Review focus area",
    "code": "Code to review"
  },
  "tags": ["review", "code"]
}
```

## Architecture

- **Storage Layer**: BadgerDB for fast key-value operations
- **Search Layer**: Bleve for full-text search and indexing  
- **MCP Layer**: JSON-RPC 2.0 protocol implementation
- **Transport Layer**: Stdio and WebSocket support
- **Daemon Layer**: Background service with HTTP endpoints

## Token Optimization

The search system is optimized for minimal token usage:

- Compact result formatting
- Relevant snippet extraction
- Score-based ranking
- Configurable result limits
- Semantic filtering

## API Endpoints (Daemon Mode)

- `GET /`: Server information page
- `GET /health`: Health check endpoint
- `WS /mcp`: WebSocket MCP endpoint

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Build for production
go build -ldflags="-s -w" -o mcp-mindport
```

## License

MIT License