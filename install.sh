#!/bin/bash

# MindPort Installation Script for Claude Code
# This script installs MindPort MCP server and configures it for Claude Desktop

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MINDPORT_VERSION="1.0.0"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/mindport"
CLAUDE_CONFIG_DIR="$HOME/Library/Application Support/Claude"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking system requirements..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+ first."
        echo "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    required_version="1.21"
    if ! printf '%s\n%s\n' "$required_version" "$go_version" | sort -V -C; then
        log_error "Go version $go_version is too old. Please upgrade to Go 1.21+"
        exit 1
    fi
    
    log_success "Go $go_version is installed"
    
    # Check if Claude Desktop is installed (macOS)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if [[ ! -d "/Applications/Claude.app" ]]; then
            log_warning "Claude Desktop not found. Please install it from the official website."
            log_info "You can still use MindPort with other MCP clients."
        else
            log_success "Claude Desktop found"
        fi
    fi
}

create_directories() {
    log_info "Creating directories..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$CONFIG_DIR/data/storage"
    mkdir -p "$CONFIG_DIR/data/search"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        mkdir -p "$CLAUDE_CONFIG_DIR"
    fi
    
    log_success "Directories created"
}

build_mindport() {
    log_info "Building MindPort..."
    
    # Get current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # Build the binary
    cd "$SCRIPT_DIR"
    go mod tidy
    go build -ldflags="-s -w" -o "$INSTALL_DIR/mcp-mindport"
    
    # Make it executable
    chmod +x "$INSTALL_DIR/mcp-mindport"
    
    log_success "MindPort built and installed to $INSTALL_DIR/mcp-mindport"
}

create_config() {
    log_info "Creating configuration..."
    
    cat > "$CONFIG_DIR/config.yaml" << EOF
server:
  host: "localhost"
  port: 8080

storage:
  path: "$CONFIG_DIR/data/storage"

search:
  index_path: "$CONFIG_DIR/data/search"

daemon:
  pid_file: "/tmp/mcp-mindport.pid"
  log_file: "$CONFIG_DIR/mindport.log"
EOF
    
    log_success "Configuration created at $CONFIG_DIR/config.yaml"
}

configure_claude_desktop() {
    if [[ "$OSTYPE" != "darwin"* ]]; then
        log_warning "Claude Desktop configuration is only supported on macOS"
        return
    fi
    
    log_info "Configuring Claude Desktop..."
    
    CLAUDE_CONFIG_FILE="$CLAUDE_CONFIG_DIR/claude_desktop_config.json"
    
    # Create backup if config exists
    if [[ -f "$CLAUDE_CONFIG_FILE" ]]; then
        cp "$CLAUDE_CONFIG_FILE" "$CLAUDE_CONFIG_FILE.backup.$(date +%s)"
        log_info "Backed up existing Claude config"
    fi
    
    # Read existing config or create new one
    if [[ -f "$CLAUDE_CONFIG_FILE" ]]; then
        existing_config=$(cat "$CLAUDE_CONFIG_FILE")
    else
        existing_config="{}"
    fi
    
    # Add MindPort to mcpServers using jq if available, otherwise use simple approach
    if command -v jq &> /dev/null; then
        echo "$existing_config" | jq '.mcpServers.mindport = {
            "command": "'$INSTALL_DIR'/mcp-mindport",
            "args": ["--config", "'$CONFIG_DIR'/config.yaml"],
            "env": {}
        }' > "$CLAUDE_CONFIG_FILE"
    else
        # Fallback method without jq
        cat > "$CLAUDE_CONFIG_FILE" << EOF
{
  "mcpServers": {
    "mindport": {
      "command": "$INSTALL_DIR/mcp-mindport",
      "args": ["--config", "$CONFIG_DIR/config.yaml"],
      "env": {}
    }
  }
}
EOF
    fi
    
    log_success "Claude Desktop configured"
    log_info "Restart Claude Desktop to load MindPort"
}

create_service_scripts() {
    log_info "Creating service scripts..."
    
    # Create start script
    cat > "$INSTALL_DIR/mindport-start" << EOF
#!/bin/bash
# Start MindPort as daemon
"$INSTALL_DIR/mcp-mindport" --daemon --config "$CONFIG_DIR/config.yaml"
EOF
    chmod +x "$INSTALL_DIR/mindport-start"
    
    # Create stop script
    cat > "$INSTALL_DIR/mindport-stop" << EOF
#!/bin/bash
# Stop MindPort daemon
if [[ -f "/tmp/mcp-mindport.pid" ]]; then
    pid=\$(cat "/tmp/mcp-mindport.pid")
    if kill -0 "\$pid" 2>/dev/null; then
        kill "\$pid"
        echo "MindPort stopped (PID: \$pid)"
    else
        echo "MindPort is not running"
        rm -f "/tmp/mcp-mindport.pid"
    fi
else
    echo "MindPort is not running"
fi
EOF
    chmod +x "$INSTALL_DIR/mindport-stop"
    
    # Create status script
    cat > "$INSTALL_DIR/mindport-status" << EOF
#!/bin/bash
# Check MindPort status
if [[ -f "/tmp/mcp-mindport.pid" ]]; then
    pid=\$(cat "/tmp/mcp-mindport.pid")
    if kill -0 "\$pid" 2>/dev/null; then
        echo "MindPort is running (PID: \$pid)"
        echo "Web interface: http://localhost:8080"
        echo "Health check: http://localhost:8080/health"
    else
        echo "MindPort is not running (stale PID file)"
        rm -f "/tmp/mcp-mindport.pid"
    fi
else
    echo "MindPort is not running"
fi
EOF
    chmod +x "$INSTALL_DIR/mindport-status"
    
    log_success "Service scripts created"
}

add_to_path() {
    log_info "Checking PATH configuration..."
    
    # Check if INSTALL_DIR is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        log_warning "$INSTALL_DIR is not in your PATH"
        
        # Add to shell profile
        shell_profile=""
        if [[ -f "$HOME/.zshrc" ]]; then
            shell_profile="$HOME/.zshrc"
        elif [[ -f "$HOME/.bashrc" ]]; then
            shell_profile="$HOME/.bashrc"
        elif [[ -f "$HOME/.bash_profile" ]]; then
            shell_profile="$HOME/.bash_profile"
        fi
        
        if [[ -n "$shell_profile" ]]; then
            echo "" >> "$shell_profile"
            echo "# MindPort MCP Server" >> "$shell_profile"
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_profile"
            log_success "Added $INSTALL_DIR to PATH in $shell_profile"
            log_info "Please run: source $shell_profile"
        else
            log_warning "Please add $INSTALL_DIR to your PATH manually"
        fi
    else
        log_success "$INSTALL_DIR is already in PATH"
    fi
}

run_tests() {
    log_info "Running tests..."
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"
    
    if go test ./internal/search/... -v; then
        log_success "All tests passed"
    else
        log_warning "Some tests failed, but installation will continue"
    fi
}

print_usage() {
    log_success "MindPort installation completed!"
    echo ""
    echo "🚀 Getting Started:"
    echo "  1. Test the installation:"
    echo "     mcp-mindport --help"
    echo ""
    echo "  2. Start as daemon:"
    echo "     mindport-start"
    echo ""
    echo "  3. Check status:"
    echo "     mindport-status"
    echo ""
    echo "  4. Stop daemon:"
    echo "     mindport-stop"
    echo ""
    echo "📖 Usage Examples:"
    echo "  # Store a resource"
    echo "  curl -X POST http://localhost:8080/mcp -d '{...}'"
    echo ""
    echo "  # Use with Claude Desktop (if configured)"
    echo "  - Restart Claude Desktop"
    echo "  - Use search tools in your conversations"
    echo ""
    echo "🔧 Configuration:"
    echo "  Config file: $CONFIG_DIR/config.yaml"
    echo "  Data directory: $CONFIG_DIR/data/"
    echo "  Logs: $CONFIG_DIR/mindport.log"
    echo ""
    echo "📚 Documentation:"
    echo "  GitHub: https://github.com/your-repo/mcp-mindport"
    echo "  MCP Docs: https://modelcontextprotocol.io/"
}

main() {
    echo "🧠 MindPort MCP Server Installer"
    echo "=================================="
    echo ""
    
    check_requirements
    create_directories
    build_mindport
    create_config
    create_service_scripts
    add_to_path
    
    # Ask about Claude Desktop configuration
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo ""
        read -p "Configure Claude Desktop integration? (y/n): " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            configure_claude_desktop
        fi
    fi
    
    # Ask about running tests
    echo ""
    read -p "Run tests to verify installation? (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        run_tests
    fi
    
    echo ""
    print_usage
}

# Handle command line arguments
case "${1:-install}" in
    "install"|"")
        main
        ;;
    "uninstall")
        log_info "Uninstalling MindPort..."
        rm -f "$INSTALL_DIR/mcp-mindport"
        rm -f "$INSTALL_DIR/mindport-start"
        rm -f "$INSTALL_DIR/mindport-stop" 
        rm -f "$INSTALL_DIR/mindport-status"
        rm -rf "$CONFIG_DIR"
        log_success "MindPort uninstalled"
        ;;
    "update")
        log_info "Updating MindPort..."
        build_mindport
        log_success "MindPort updated"
        ;;
    "test")
        run_tests
        ;;
    "help"|"-h"|"--help")
        echo "MindPort Installer"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  install    Install MindPort (default)"
        echo "  uninstall  Remove MindPort"
        echo "  update     Update MindPort"
        echo "  test       Run tests"
        echo "  help       Show this help"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Run '$0 help' for usage information"
        exit 1
        ;;
esac