# Quick Start Guide

Get MindPort running in 2 minutes on a fresh machine!

## Prerequisites

- **Node.js 18+** - [Download here](https://nodejs.org/)
- **npm** (comes with Node.js)

## One-Command Setup

```bash
# Clone and setup everything
git clone <repository-url>
cd mcp-mindport
./setup.sh
```

That's it! The setup script will:
- ✅ Install all dependencies (main + web interface)
- ✅ Run tests to verify everything works
- ✅ Configure Claude Desktop integration (if detected)
- ✅ Show you next steps

## Manual Setup (if preferred)

```bash
# 1. Install main dependencies
npm install

# 2. Install web interface dependencies  
cd site && npm install && cd ..

# 3. Run tests
npm test
```

## Usage

### Start MCP Server (for Claude Desktop)
```bash
npm run dev
```

### Start Web Interface
```bash
npm run web
# Visit http://localhost:3001
```

### Run Tests
```bash
npm test          # Watch mode
npm run test:run  # Run once
npm run test:ui   # Interactive UI
```

## What You Get

### 🤖 MCP Server (14 Tools)
- **Resource Management**: store_resource, get_resource, list_resources
- **Advanced Search**: search_resources, advanced_search, grep, find
- **Domain Management**: create_domain, list_domains, switch_domain
- **Prompt Templates**: store_prompt, list_prompts, get_prompt

### 🌐 Web Interface
- **Dashboard** with statistics and recent activity
- **Resource Browser** with search and filtering
- **Prompt Template Manager** with variable testing
- **Domain Explorer** for organization

### 🧪 Test Suite
- **76 comprehensive tests** covering all functionality
- **Performance benchmarks** for large datasets
- **Integration tests** for end-to-end workflows

## Troubleshooting

### "Command not found: ts-node"
```bash
npm install -g ts-node
# Or use: npx ts-node index.ts
```

### "Cannot find module"
```bash
# Reinstall dependencies
rm -rf node_modules site/node_modules
npm run setup
```

### "Database not found" (web interface)
```bash
# Create database by running MCP server first
npm run dev
# Then in another terminal:
npm run web
```

### Claude Desktop Integration
1. Run setup script: `./setup.sh`
2. Restart Claude Desktop
3. MindPort tools should appear in Claude

## Next Steps

1. **Store your first resource**:
   ```bash
   npm run dev  # Start MCP server
   # Use Claude Desktop to store content
   ```

2. **Browse via web interface**:
   ```bash
   npm run web
   # Visit http://localhost:3001
   ```

3. **Explore the codebase**:
   - `src/` - TypeScript MCP server
   - `site/` - Next.js web interface  
   - `tests/` - Comprehensive test suite
   - `docs/` - API and testing documentation

**Happy building!** 🚀

For detailed documentation, see [README.md](README.md)