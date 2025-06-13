#!/bin/bash

# MindPort Setup Script
# Sets up MindPort MCP Server for fresh machine installation

set -e

echo "🚀 Setting up MindPort MCP Server..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed!"
    echo "Please install Node.js (v18 or higher) from https://nodejs.org/"
    echo "Or use a package manager:"
    echo "  macOS: brew install node"
    echo "  Ubuntu: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    print_warning "Node.js version $NODE_VERSION detected. Version 18+ recommended."
fi

print_status "Node.js $(node --version) found ✓"

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    print_error "npm is not installed!"
    exit 1
fi

print_status "npm $(npm --version) found ✓"

# Install main project dependencies
print_status "Installing MCP server dependencies..."
npm install

if [ $? -eq 0 ]; then
    print_success "MCP server dependencies installed"
else
    print_error "Failed to install MCP server dependencies"
    exit 1
fi

# Install web interface dependencies
print_status "Installing web interface dependencies..."
cd site
npm install

if [ $? -eq 0 ]; then
    print_success "Web interface dependencies installed"
else
    print_error "Failed to install web interface dependencies"
    exit 1
fi

cd ..

# Build the project
print_status "Building project..."
npm run build

if [ $? -eq 0 ]; then
    print_success "Project built successfully"
else
    print_error "Failed to build project"
    exit 1
fi

# Run tests to verify everything works
print_status "Running tests to verify installation..."
npm run test:run

if [ $? -eq 0 ]; then
    print_success "All tests passed ✓"
else
    print_warning "Some tests failed, but installation can continue"
fi

# Test MCP server startup (brief)
print_status "Testing MCP server startup..."
timeout 3s node dist/index.js --version 2>/dev/null || print_status "MCP server test completed"

# Setup Claude Desktop integration (optional)
CLAUDE_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
if [ -d "$HOME/Library/Application Support/Claude" ]; then
    print_status "Claude Desktop detected. Setting up integration..."
    
    # Backup existing config
    if [ -f "$CLAUDE_CONFIG" ]; then
        cp "$CLAUDE_CONFIG" "$CLAUDE_CONFIG.backup.$(date +%s)"
        print_status "Existing Claude Desktop config backed up"
    fi
    
    # Create config directory if it doesn't exist
    mkdir -p "$(dirname "$CLAUDE_CONFIG")"
    
    # Get current directory
    CURRENT_DIR="$(pwd)"
    
    # Create or update Claude Desktop config
    cat > "$CLAUDE_CONFIG" << EOF
{
  "mcpServers": {
    "mindport": {
      "command": "node",
      "args": ["$CURRENT_DIR/dist/index.js"],
      "env": {
        "MCP_MINDPORT_LOG": "discard"
      }
    }
  }
}
EOF
    
    print_success "Claude Desktop integration configured!"
    print_warning "Please restart Claude Desktop to activate MindPort"
else
    print_status "Claude Desktop not found - you can set up integration later"
fi

echo
echo "🎉 MindPort setup complete!"
echo
echo "📋 Quick Start Commands:"
echo "  npm run dev        # Start MCP server (for Claude Desktop)"
echo "  npm run web        # Start web interface (http://localhost:3001)"
echo "  npm test           # Run test suite"
echo "  npm run test:ui    # Run tests with UI"
echo
echo "📚 Usage:"
echo "  1. Start the MCP server: npm run dev"
echo "  2. Use Claude Desktop to interact with MindPort tools"
echo "  3. Browse data via web interface: npm run web"
echo
echo "🔧 Available MCP Tools:"
echo "  • store_resource    - Store content with metadata"
echo "  • search_resources  - Fuzzy search across content"
echo "  • get_resource      - Retrieve specific resources"
echo "  • list_resources    - List resources in domain"
echo "  • advanced_search   - Complex queries with filters"
echo "  • grep             - Regex pattern matching"
echo "  • find             - Name-based resource discovery"
echo "  • list_domains     - List available domains"
echo "  • create_domain    - Create new domain contexts"
echo "  • switch_domain    - Change current domain"
echo "  • store_prompt     - Store reusable prompt templates"
echo "  • list_prompts     - List available prompts"
echo "  • get_prompt       - Retrieve and render prompts"
echo
echo "📖 Documentation: See README.md for full documentation"
echo "🐛 Issues: Report at https://github.com/your-repo/issues"
echo

if [ -d "$HOME/Library/Application Support/Claude" ]; then
    echo "💡 Don't forget to restart Claude Desktop!"
fi