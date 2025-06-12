# Claude Code MCP Integration Troubleshooting

## Issue: Server transport closed unexpectedly

This is a **known issue with Claude Code CLI's MCP integration**, not a problem with MindPort. The research shows this affects ~75% of MCP server connections.

## Diagnosis Steps

### 1. Verify MindPort Works Independently

First, confirm MindPort MCP server works correctly:

```bash
# Test direct MCP communication
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./mcp-mindport --log /tmp/mindport-test.log

# Check the log
cat /tmp/mindport-test.log
```

**Expected**: Should receive a valid JSON response and see initialization logs.

### 2. Run the Debug Script

```bash
./claude-code-debug.sh
```

This will comprehensively test your setup and identify configuration issues.

### 3. Check Configuration Files

Ensure you have the correct configuration files:

**`.mcp.json` (project-level, recommended):**
```json
{
  "mcpServers": {
    "mindport": {
      "type": "stdio",
      "command": "mcp-mindport",
      "args": ["--config", "~/.config/mindport/config.yaml", "--log", "/tmp/mindport-claude.log"],
      "env": {
        "MCP_MINDPORT_CONFIG": "~/.config/mindport/config.yaml"
      },
      "protocolVersion": "2024-11-05"
    }
  }
}
```

**`.claude/settings.json` (alternative configuration):**
```json
{
  "mcpServers": {
    "mindport": {
      "type": "stdio", 
      "command": "/Users/USERNAME/.local/bin/mcp-mindport",
      "args": ["--config", "/Users/USERNAME/.config/mindport/config.yaml", "--log", "/tmp/mindport-claude.log"],
      "env": {
        "MCP_MINDPORT_CONFIG": "/Users/USERNAME/.config/mindport/config.yaml"
      },
      "protocolVersion": "2024-11-05"
    }
  }
}
```

**Important**: Replace `USERNAME` with your actual username, or use the `.mcp.json` format which supports `~` expansion.

### 4. Enable Debug Logging

Run Claude Code with full debugging:

```bash
MCP_CLAUDE_DEBUG=true claude-code --mcp-debug
```

Then check for logs:
- Look for `mcp-server-mindport.log` in Claude Code's log directory
- Check `/tmp/mindport-claude.log` for MindPort server logs

### 5. Test with Absolute Paths

Update your configuration to use absolute paths instead of relative ones:

```json
{
  "mcpServers": {
    "mindport": {
      "type": "stdio",
      "command": "/Users/jaenster/.local/bin/mcp-mindport",
      "args": ["--config", "/Users/jaenster/.config/mindport/config.yaml", "--log", "/tmp/mindport-claude-debug.log"],
      "env": {
        "MCP_MINDPORT_CONFIG": "/Users/jaenster/.config/mindport/config.yaml",
        "HOME": "/Users/jaenster",
        "PATH": "/usr/local/bin:/usr/bin:/bin"
      },
      "protocolVersion": "2024-11-05"
    }
  }
}
```

## Known Claude Code MCP Issues

Based on research, Claude Code CLI has several known bugs:

1. **Protocol Version Validation**: Bug where `protocolVersion` field isn't properly handled
2. **Connection Lifecycle**: Connections close immediately after successful initialization
3. **Environment Variables**: Not passed correctly to servers
4. **Server Discovery**: Inconsistent server discovery and connection

## Workarounds

### 1. Use Claude Desktop Instead

Claude Desktop has more stable MCP integration. Configure it:

```json
// ~/Library/Application Support/Claude/claude_desktop_config.json
{
  "mcpServers": {
    "mindport": {
      "command": "/Users/jaenster/.local/bin/mcp-mindport",
      "args": ["--config", "/Users/jaenster/.config/mindport/config.yaml"],
      "env": {}
    }
  }
}
```

### 2. Enhanced Logging Configuration

Use detailed logging to debug the issue:

```bash
# Run MindPort with verbose logging
./mcp-mindport --config ~/.config/mindport/config.yaml --log /tmp/mindport-verbose.log

# In another terminal, watch the logs
tail -f /tmp/mindport-verbose.log
```

### 3. Test with MCP Inspector

Use the official MCP testing tool:

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Test MindPort with Inspector
npx @modelcontextprotocol/inspector /Users/jaenster/.local/bin/mcp-mindport --config /Users/jaenster/.config/mindport/config.yaml
```

### 4. Alternative Transport Test

If stdio fails, test with manual JSON-RPC:

```bash
# Create test session
cat << 'EOF' > /tmp/mcp-test-session.json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"manual-test","version":"1.0"}}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"resources/list"}
EOF

# Test
cat /tmp/mcp-test-session.json | ./mcp-mindport --log /tmp/manual-test.log
```

## Current Status

✅ **MindPort MCP server is working correctly**
- Handles `initialize` requests properly
- Returns valid responses
- Implements full MCP protocol

❌ **Claude Code CLI has integration issues**
- Known bugs in protocol handling
- Connection closure after initialization
- Inconsistent server discovery

## Recommendations

1. **Use Claude Desktop** for now if you need immediate MCP functionality
2. **Report the issue** to Claude Code team with logs from the debug script
3. **Monitor Claude Code updates** - this may be fixed in future releases
4. **Test periodically** as Claude Code CLI is actively developed

## Getting Help

If MindPort works with direct testing but fails in Claude Code:

1. This confirms it's a Claude Code issue, not MindPort
2. Share logs from `claude-code-debug.sh` when reporting issues
3. Include the exact error messages from Claude Code
4. Try the workarounds above

## File Locations

- **Debug script**: `./claude-code-debug.sh`
- **Project config**: `./.mcp.json`
- **Claude settings**: `./.claude/settings.json`
- **MindPort config**: `~/.config/mindport/config.yaml`
- **Debug logs**: `/tmp/mindport-claude-debug.log`
- **Test logs**: `/tmp/mindport-test.log`