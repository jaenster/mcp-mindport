# MindPort Installation Guide

## Quick Install

### One-Line Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/your-repo/mcp-mindport/main/install.sh | bash
```

### Manual Install
```bash
git clone https://github.com/your-repo/mcp-mindport.git
cd mcp-mindport
chmod +x install.sh
./install.sh
```

## System Requirements

- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **macOS/Linux** - Windows support via WSL
- **Claude Desktop** (optional) - For MCP integration

## Installation Options

### 1. Claude Desktop Integration (Recommended)

The installer will automatically configure Claude Desktop to use MindPort:

```bash
./install.sh
# Follow prompts to configure Claude Desktop
```

**Manual Claude Desktop Configuration:**

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mindport": {
      "command": "/Users/[username]/.local/bin/mcp-mindport",
      "args": ["--config", "/Users/[username]/.config/mindport/config.yaml"],
      "env": {}
    }
  }
}
```

### 2. Claude Code CLI Integration

For Claude Code CLI, add MindPort as an MCP server:

```bash
# Using environment variables
export MCP_MINDPORT_CONFIG="$HOME/.config/mindport/config.yaml"

# Or specify in your .zshrc/.bashrc
echo 'export MCP_MINDPORT_CONFIG="$HOME/.config/mindport/config.yaml"' >> ~/.zshrc
```

**Use with Claude Code:**
```bash
# Start MindPort daemon
mindport-start

# Use Claude Code with MCP server
claude-code --mcp-server mindport
```

### 3. Standalone Usage

MindPort supports two modes:

**Stdio Mode (Default - Standard MCP):**
```bash
# Run in stdio mode for direct MCP communication (logging disabled by default)
mcp-mindport

# With custom config (logging still disabled by default)
mcp-mindport --config /path/to/config.yaml

# In specific domain (logging still disabled by default)
mcp-mindport --domain project1

# Enable logging to file in stdio mode
mcp-mindport --log /path/to/mindport.log

# Force logging to stderr in stdio mode (not recommended)
mcp-mindport --log stderr
```

**Daemon Mode (Optional - WebSocket support):**
```bash
# Start as daemon with WebSocket endpoint
mcp-mindport --daemon

# Access via WebSocket at ws://localhost:8080/mcp
# Or HTTP endpoints at http://localhost:8080/
```

## Configuration

Default configuration is created at `~/.config/mindport/config.yaml`:

```yaml
server:
  host: "localhost"
  port: 8080

storage:
  path: "~/.config/mindport/data/storage"

search:
  index_path: "~/.config/mindport/data/search"

domain:
  default_domain: "default"
  current_domain: ""

daemon:
  pid_file: "/tmp/mcp-mindport.pid"
  log_file: "~/.config/mindport/mindport.log"
```

### Environment Variables

Override config with environment variables:

```bash
export MCP_MINDPORT_CONFIG="/path/to/config.yaml"
export MCP_MINDPORT_DAEMON="true"
export MCP_MINDPORT_DOMAIN="project1"
export MCP_MINDPORT_DEFAULT_DOMAIN="work"
export MCP_MINDPORT_LOG="discard"  # Disable logging for stdio mode
```

## Usage Examples

### Basic Resource Storage
```bash
# Store a code snippet
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "store_resource",
      "arguments": {
        "title": "React Hook Example",
        "content": "const [count, setCount] = useState(0);",
        "type": "code",
        "tags": ["react", "hooks", "javascript"]
      }
    }
  }'
```

### Search Resources
```bash
# Search for React hooks
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "search_resources",
      "arguments": {
        "query": "React hooks",
        "limit": 5
      }
    }
  }'
```

### Domain Management
```bash
# Create a project domain
mcp-mindport --domain project1 --create-domain

# List domains
mcp-mindport --list-domains

# Start in specific domain
mcp-mindport --domain project1
```

## Service Management

The installer creates convenience scripts:

```bash
# Start MindPort daemon
mindport-start

# Check status
mindport-status

# Stop daemon
mindport-stop
```

### Manual Service Management
```bash
# Start with custom config
mcp-mindport --daemon --config /path/to/config.yaml

# Start in specific domain
mcp-mindport --daemon --domain project1

# Stop by PID
kill $(cat /tmp/mcp-mindport.pid)
```

## Verification

### Test Installation
```bash
# Check binary
mcp-mindport --help

# Test daemon mode
mindport-start
mindport-status
curl http://localhost:8080/health
mindport-stop
```

### Test MCP Integration
```bash
# Test with MCP client
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}' | mcp-mindport
```

### Run Test Suite
```bash
cd mcp-mindport
go test ./...
# or
./install.sh test
```

## Troubleshooting

### Common Issues

**1. Permission Denied**
```bash
chmod +x ~/.local/bin/mcp-mindport
```

**2. Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Use different port
export MCP_MINDPORT_PORT=8081
```

**3. Claude Desktop Not Finding Server**
- Restart Claude Desktop completely
- Check config file syntax: `cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | jq`
- Verify binary path: `which mcp-mindport`

**4. Go Build Errors**
```bash
# Update Go modules
cd mcp-mindport
go mod tidy
go mod download
```

### Logs and Debugging

```bash
# Check daemon logs
tail -f ~/.config/mindport/mindport.log

# Check Claude Desktop logs (macOS)
tail -f ~/Library/Logs/Claude/mcp.log

# Run with debug output
mcp-mindport --daemon --verbose
```

### Reset Installation
```bash
# Uninstall completely
./install.sh uninstall

# Clean reinstall
rm -rf ~/.config/mindport
./install.sh
```

## Advanced Configuration

### Custom Domains
```yaml
domain:
  default_domain: "work"
  domains:
    work:
      name: "Work Projects"
      description: "Work-related resources"
    personal:
      name: "Personal Projects"
      description: "Personal coding projects"
```

### Performance Tuning
```yaml
storage:
  batch_size: 1000
  sync_writes: false

search:
  max_results: 100
  highlight_fragments: 3
```

### Security Settings
```yaml
server:
  enable_cors: false
  allowed_origins: ["localhost"]
  rate_limit: 100
```

## Integration Examples

### With VS Code
Create a VS Code task in `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Start MindPort",
      "type": "shell",
      "command": "mindport-start",
      "group": "build"
    }
  ]
}
```

### With Alfred/Raycast
Create workflow to search MindPort resources directly from launcher.

### With Git Hooks
Auto-index code changes:
```bash
#!/bin/sh
# .git/hooks/post-commit
mcp-mindport-index-diff HEAD~1..HEAD
```

## Uninstallation

```bash
# Complete removal
./install.sh uninstall

# Manual cleanup
rm -rf ~/.config/mindport
rm -f ~/.local/bin/mcp-mindport
rm -f ~/.local/bin/mindport-*
```

## Support

- **Documentation**: [GitHub Wiki](https://github.com/your-repo/mcp-mindport/wiki)
- **Issues**: [GitHub Issues](https://github.com/your-repo/mcp-mindport/issues)
- **MCP Protocol**: [Official Docs](https://modelcontextprotocol.io/)